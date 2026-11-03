package runtime

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MontFerret/api"
)

func TestHTTPOutputCancellationUnblocksRead(t *testing.T) {
	for _, mode := range []string{"invocation", "consumer", "close"} {
		t.Run(mode, func(t *testing.T) {
			invocation, cancelInvocation := context.WithCancel(t.Context())
			defer cancelInvocation()
			consumption, cancelConsumption := context.WithCancel(t.Context())
			defer cancelConsumption()
			reader, writer := io.Pipe()
			defer writer.Close()
			body := &responseBody{reader: reader}
			response := &http.Response{Body: body, ContentLength: -1}
			out := newHTTPOutput(invocation, response, func() {})
			entered := make(chan struct{})
			done := make(chan error, 1)
			go func() {
				done <- out.Consume(consumption, func(_ context.Context, chunk []byte) error {
					if string(chunk) != "prefix" {
						t.Errorf("unexpected bytes %q", chunk)
					}

					close(entered)

					return nil
				})
			}()

			if _, err := io.WriteString(writer, "prefix"); err != nil {
				t.Fatal(err)
			}

			<-entered
			want := error(context.Canceled)
			switch mode {
			case "consumer":
				cancelConsumption()
			case "invocation":
				cancelInvocation()
			case "close":
				want = api.ErrOutputClosed
				go func() { _ = out.Close() }()
			}

			select {
			case err := <-done:
				if !errors.Is(err, want) {
					t.Fatalf("consume returned %v, want %v", err, want)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("cancellation did not unblock the response body")
			}

			if err := out.Close(); err != nil || body.closes != 1 {
				t.Fatalf("close=%v, body closes=%d", err, body.closes)
			}
		})
	}
}

func TestHTTPOutputCloseWaitsForConsumer(t *testing.T) {
	body := &responseBody{reader: strings.NewReader("bytes")}
	out := newHTTPOutput(t.Context(), &http.Response{Body: body, ContentLength: 5}, func() {})
	entered, stopping, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	releaseCallback := sync.OnceFunc(func() { close(release) })
	defer releaseCallback()

	consumed := make(chan error, 1)
	go func() {
		consumed <- out.Consume(t.Context(), func(ctx context.Context, chunk []byte) error {
			close(entered)
			<-ctx.Done()
			close(stopping)
			<-release

			if string(chunk) != "bytes" {
				t.Error("borrowed chunk changed before callback returned")
			}

			return nil
		})
	}()
	<-entered

	if _, err := out.Collect(t.Context()); !errors.Is(err, api.ErrOutputInUse) {
		t.Fatalf("competing collection: %v", err)
	}

	closed := make(chan error, 1)
	go func() { closed <- out.Close() }()

	select {
	case <-stopping:
	case <-time.After(2 * time.Second):
		t.Fatal("callback did not observe close")
	}

	select {
	case <-closed:
		t.Fatal("close returned during callback")
	default:
	}

	releaseCallback()
	if err := <-consumed; !errors.Is(err, api.ErrOutputClosed) {
		t.Fatalf("interrupted consumption: %v", err)
	}

	if err := <-closed; err != nil || body.closes != 1 {
		t.Fatalf("close=%v, body closes=%d", err, body.closes)
	}
}

func TestHTTPOutputMetadataAndAdmission(t *testing.T) {
	body := &responseBody{reader: strings.NewReader("abc")}
	out := newHTTPOutput(t.Context(), &http.Response{Body: body, ContentLength: 4}, func() {})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := out.Collect(ctx); !errors.Is(err, context.Canceled) || body.closes != 0 {
		t.Fatalf("rejected consumption claimed body: %v", err)
	}

	metadata := out.Metadata()
	content, err := out.Collect(t.Context())
	if !errors.Is(err, io.ErrUnexpectedEOF) || content == nil || string(content.Data) != "abc" ||
		content.Metadata != metadata || !metadata.LengthKnown || metadata.Length != 4 || body.closes != 1 {
		t.Fatalf("truncated output: %#v, %v, closes=%d", content, err, body.closes)
	}

	if _, err := out.Collect(t.Context()); !errors.Is(err, api.ErrOutputClosed) {
		t.Fatalf("reuse after finalization: %v", err)
	}
}

func TestHTTPOutputAbandonBeforeConsumption(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		released := make(chan struct{})
		closeErr := errors.New("close failure")
		body := &responseBody{closeErr: closeErr}
		out := newHTTPOutput(ctx, &http.Response{Body: body}, func() { close(released) })
		if explicit {
			if err := out.Close(); !errors.Is(err, closeErr) {
				t.Fatalf("abandon: %v", err)
			}
		}

		cancel()
		select {
		case <-released:
		case <-time.After(2 * time.Second):
			t.Fatal("unclaimed body remained open")
		}

		if err := out.Close(); !errors.Is(err, closeErr) || body.closes != 1 {
			t.Fatalf("retained close=%v, closes=%d", err, body.closes)
		}
	}
}
