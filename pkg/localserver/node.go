package localserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

type (
	NodeSettings struct {
		Name          string
		Port          int
		Handler       http.Handler
		BindHost      string
		AdvertiseHost string
	}

	Node struct {
		id       int
		settings NodeSettings
		server   *http.Server
		listener net.Listener
		running  atomic.Bool
		mu       sync.RWMutex
		serveErr error
	}
)

var nodeIDCounter atomic.Int64

func NewNode(settings NodeSettings) (*Node, error) {
	if settings.Handler == nil {
		return nil, errors.New("local server handler is required")
	}

	return &Node{
		id:       int(nodeIDCounter.Add(1)),
		settings: settings,
		server: &http.Server{
			Handler: settings.Handler,
		},
	}, nil
}

func (n *Node) ID() int {
	return n.id
}

func (n *Node) Name() string {
	return n.settings.Name
}

func (n *Node) Port() int {
	return n.settings.Port
}

func (n *Node) IsRunning() bool {
	return n.running.Load()
}

func (n *Node) ServeErr() error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.serveErr
}

func (n *Node) ListenerAddr() net.Addr {
	if n.listener == nil {
		return nil
	}

	return n.listener.Addr()
}

func (n *Node) Start(ctx context.Context) error {
	if n.running.Load() {
		return nil
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", net.JoinHostPort(n.settings.BindHost, fmt.Sprintf("%d", n.settings.Port)))
	if err != nil {
		return err
	}

	if addr, ok := listener.Addr().(*net.TCPAddr); ok {
		n.settings.Port = addr.Port
	}

	n.listener = listener
	n.running.Store(true)

	go func() {
		if err := n.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			n.mu.Lock()
			n.serveErr = err
			n.mu.Unlock()
			n.running.Store(false)
		}
	}()

	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	n.running.Store(false)
	return n.server.Shutdown(ctx)
}

func (n *Node) String() string {
	return EndpointURL(n.settings.AdvertiseHost, n.Port())
}
