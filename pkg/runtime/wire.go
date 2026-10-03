package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/MontFerret/api"
	"github.com/MontFerret/ferret/v2"
	wireclient "github.com/MontFerret/wire/client"
)

type (
	wireRuntime struct {
		remote    api.Runtime
		params    map[string]any
		transport *wireTransport
		once      sync.Once
		err       error
	}

	// wireTransport shares one retained close between construction cancellation and cleanup.
	wireTransport struct {
		connection io.Closer
		once       sync.Once
		err        error
	}
)

func newWire(ctx context.Context, opts Options) (*wireRuntime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	address, err := wireAddress(opts.Endpoint)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, opts.ConnectTimeout)
	defer cancel()
	var dialed atomic.Bool

	connection, err := grpc.NewClient("passthrough:///"+address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithNoProxy(),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			if !dialed.CompareAndSwap(false, true) {
				return nil, errors.New("wire transport connection was lost; start a new command")
			}

			return new(net.Dialer).DialContext(ctx, "tcp4", address)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("connect Wire runtime %s: %w", opts.Endpoint, err)
	}

	physical := &wireTransport{connection: connection}
	// Wire detaches its Connect stream from the setup context. Closing the
	// transport also bounds stream creation when the HTTP/2 handshake stalls.
	closed := make(chan struct{})
	stop := context.AfterFunc(connectCtx, func() {
		_ = physical.Close()
		close(closed)
	})
	remote, err := wireclient.New(connectCtx, connection)
	if !stop() {
		<-closed
	}

	if ctxErr := connectCtx.Err(); ctxErr != nil {
		err = errors.Join(err, ctxErr)
	}

	rt := &wireRuntime{remote: remote, params: maps.Clone(opts.Params), transport: physical}
	if err != nil {
		return nil, errors.Join(fmt.Errorf("connect Wire runtime %s: %w", opts.Endpoint, err), rt.Close())
	}

	return rt, nil
}

func (rt *wireRuntime) Version(ctx context.Context) (string, error) {
	value, err := rt.remote.Version(ctx)

	return string(value), err
}

func (rt *wireRuntime) Run(ctx context.Context, query ferret.Source, params map[string]any) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	shared, err := wireParameters(rt.params)
	if err != nil {
		return nil, err
	}

	perRun, err := wireParameters(params)
	if err != nil {
		return nil, err
	}

	out, err := rt.remote.Run(ctx,
		api.Source{Name: query.Name(), Content: query.Content()},
		api.WithParams(shared), api.WithParams(perRun),
	)
	if out == nil {
		return nil, err
	}

	return out.Content, err
}

func (rt *wireRuntime) Close() error {
	rt.once.Do(func() {
		if rt.remote != nil {
			rt.err = rt.remote.Close()
		}

		rt.err = errors.Join(rt.err, rt.transport.Close())
	})

	return rt.err
}

func (t *wireTransport) Close() error {
	t.once.Do(func() { t.err = t.connection.Close() })

	return t.err
}
