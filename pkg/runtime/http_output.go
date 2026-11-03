package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/MontFerret/api"
)

type (
	httpOutput struct {
		invocation context.Context
		metadata   api.Metadata
		body       io.ReadCloser
		closeBody  func() error

		mu             sync.Mutex
		state          httpOutputState
		explicitClose  bool
		operation      context.Context
		cancel         context.CancelFunc
		stopInvocation func() bool
		done           chan struct{}
	}

	httpOutputState uint8
)

const (
	httpOutputReady httpOutputState = iota
	httpOutputReading
	httpOutputStopping
	httpOutputFinalizing
)

var httpOutputBuffers = sync.Pool{New: func() any { return new([32 * 1024]byte) }}

func newHTTPOutput(ctx context.Context, response *http.Response, cancelRequest context.CancelFunc) *httpOutput {
	o := &httpOutput{
		invocation: ctx,
		body:       response.Body,
		metadata:   api.Metadata{ContentType: response.Header.Get("Content-Type")},
		done:       make(chan struct{}),
	}

	if response.ContentLength >= 0 {
		o.metadata.LengthKnown = true
		o.metadata.Length = response.ContentLength
	}

	o.closeBody = sync.OnceValue(func() error {
		cancelRequest()

		if err := o.body.Close(); err != nil {
			return fmt.Errorf("close response data: %w", err)
		}

		return nil
	})

	o.mu.Lock()
	o.stopInvocation = context.AfterFunc(ctx, func() { o.stop(false) })
	o.mu.Unlock()

	return o
}

func (o *httpOutput) Metadata() api.Metadata { return o.metadata }

func (o *httpOutput) Consume(ctx context.Context, consumer api.Consumer) (err error) {
	if ctx == nil || consumer == nil {
		return errors.New("output consumption requires a context and consumer")
	}

	o.mu.Lock()

	if err := ctx.Err(); err != nil {
		o.mu.Unlock()

		return err
	}

	if err := o.invocation.Err(); err != nil {
		o.mu.Unlock()

		return err
	}

	switch o.state {
	case httpOutputReading:
		active := o.operation.Err() == nil
		o.mu.Unlock()

		if active {
			return api.ErrOutputInUse
		}

		return api.ErrOutputClosed
	case httpOutputStopping, httpOutputFinalizing:
		o.mu.Unlock()

		return api.ErrOutputClosed
	}

	var operation context.Context
	var cancel context.CancelFunc
	if deadline, ok := o.invocation.Deadline(); ok {
		operation, cancel = context.WithDeadline(ctx, deadline)
	} else {
		operation, cancel = context.WithCancel(ctx)
	}

	o.operation, o.cancel = operation, cancel
	o.state = httpOutputReading
	o.mu.Unlock()

	stopRead := context.AfterFunc(operation, func() { o.stop(false) })
	defer func() {
		o.mu.Lock()
		// Commit the outcome before canceling our own contexts or closing the body.
		for _, cause := range []error{ctx.Err(), o.invocation.Err(), operation.Err()} {
			if cause != nil && !errors.Is(err, cause) {
				err = errors.Join(err, cause)
			}
		}

		if o.explicitClose {
			err = errors.Join(err, api.ErrOutputClosed)
		}

		o.state = httpOutputFinalizing
		o.mu.Unlock()
		stopRead()
		cancel()

		if closeErr := o.finalize(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	return o.read(operation, consumer)
}

func (o *httpOutput) read(ctx context.Context, consumer api.Consumer) error {
	buffer := httpOutputBuffers.Get().(*[32 * 1024]byte)
	defer httpOutputBuffers.Put(buffer)

	var total int64
	var emptyReads int
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		var failure error

		n, readErr := o.body.Read(buffer[:])
		if readErr != nil && readErr != io.EOF {
			failure = fmt.Errorf("read response data: %w", readErr)
		}

		if n > 0 {
			total += int64(n)
			emptyReads = 0

			if err := ctx.Err(); err != nil {
				return errors.Join(err, failure)
			}

			if err := consumer(ctx, buffer[:n:n]); err != nil {
				return errors.Join(err, failure)
			}
		}

		if failure != nil {
			return failure
		}

		if readErr == io.EOF {
			if o.metadata.LengthKnown && total != o.metadata.Length {
				return fmt.Errorf("read response data: content length mismatch: %w", io.ErrUnexpectedEOF)
			}

			if total == 0 {
				return consumer(ctx, buffer[:0:0])
			}

			return nil
		}

		if n == 0 {
			emptyReads++
			if emptyReads >= 100 {
				return io.ErrNoProgress
			}
		}
	}
}

func (o *httpOutput) Collect(ctx context.Context) (*api.Content, error) {
	var content *api.Content
	err := o.Consume(ctx, func(_ context.Context, chunk []byte) error {
		if content == nil {
			content = &api.Content{Metadata: o.metadata, Data: []byte{}}
		}

		content.Data = append(content.Data, chunk...)

		return nil
	})

	return content, err
}

func (o *httpOutput) Close() error {
	o.stop(true)
	<-o.done

	return o.closeBody()
}

// An active consumer owns finalization. Interrupt its read, but keep its chunk
// valid until the callback returns, including when Close races cancellation.
func (o *httpOutput) stop(explicit bool) {
	o.mu.Lock()
	switch o.state {
	case httpOutputReady:
		o.state = httpOutputFinalizing
		o.mu.Unlock()
		o.finalize()
	case httpOutputReading, httpOutputStopping:
		o.explicitClose = o.explicitClose || explicit
		o.state = httpOutputStopping
		o.cancel()
		o.mu.Unlock()
		o.closeBody()
	default:
		o.mu.Unlock()
	}
}

func (o *httpOutput) finalize() error {
	o.stopInvocation()
	err := o.closeBody()
	o.body = nil
	close(o.done)

	return err
}
