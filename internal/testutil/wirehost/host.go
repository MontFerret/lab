package wirehost

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/wire/server"
)

type (
	// Host owns its listener and server, while borrowing the supplied runtime.
	Host struct {
		Endpoint string
		server   *server.Server
		listener *listener
		served   chan error
		once     sync.Once
		err      error
		mu       sync.Mutex
		accepted int
		active   int
		changed  chan struct{}
	}

	listener struct {
		net.Listener
		host *Host
	}

	connection struct {
		net.Conn
		host *Host
		once sync.Once
		err  error
	}
)

// New serves a controlled runtime over an ephemeral IPv4 loopback port.
func New(t testing.TB, runtime api.Runtime) *Host {
	t.Helper()
	h := &Host{served: make(chan error, 1)}
	t.Cleanup(func() {
		if err := h.Close(); err != nil {
			t.Errorf("close Wire host: %v", err)
		}
	})

	var err error
	h.server, err = server.NewServer(runtime)
	if err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	h.listener = &listener{Listener: ln, host: h}
	h.Endpoint = "tcp://" + ln.Addr().String()
	go func() { h.served <- h.server.Serve(context.Background(), h.listener) }()

	return h
}

// Connections counts physical connections, including those already closed.
func (h *Host) Connections() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.accepted
}

// WaitForClientsClosed observes transport release without closing clients itself.
func (h *Host) WaitForClientsClosed(ctx context.Context) error {
	for {
		h.mu.Lock()
		if h.active == 0 {
			h.mu.Unlock()

			return nil
		}

		if h.changed == nil {
			h.changed = make(chan struct{})
		}

		changed := h.changed
		h.mu.Unlock()

		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Close shuts down Wire and its listener without closing the borrowed runtime.
func (h *Host) Close() error {
	h.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if h.server != nil {
			h.err = h.server.Shutdown(ctx)
		}

		if h.listener != nil {
			if err := h.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				h.err = errors.Join(h.err, err)
			}

			select {
			case err := <-h.served:
				h.err = errors.Join(h.err, err)
			case <-ctx.Done():
				h.err = errors.Join(h.err, ctx.Err())
			}
		}
	})

	return h.err
}

func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	l.host.mu.Lock()
	l.host.accepted++
	l.host.active++
	l.host.mu.Unlock()

	return &connection{Conn: conn, host: l.host}, nil
}

func (c *connection) Close() error {
	c.once.Do(func() {
		c.err = c.Conn.Close()
		c.host.mu.Lock()
		defer c.host.mu.Unlock()
		c.host.active--

		if c.host.changed != nil {
			close(c.host.changed)
			c.host.changed = nil
		}
	})

	return c.err
}
