package staticserver

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	NodeSettings struct {
		Name   string
		Port   int
		Dir    string
		Prefix string
	}

	Node struct {
		id       int
		settings NodeSettings
		engine   *echo.Echo
		running  bool
	}
)

func NewNode(settings NodeSettings) (*Node, error) {
	engine := echo.New()
	engine.Debug = false
	engine.HideBanner = true
	engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodOptions},
		AllowHeaders: []string{"*"},
	}))

	prefix := "/"
	if settings.Prefix != "" {
		prefix = settings.Prefix
	}

	engine.Static(prefix, settings.Dir)

	return &Node{
		id:       rand.Int(),
		settings: settings,
		engine:   engine,
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
	return n.running
}

func (n *Node) Start(_ context.Context) error {
	if n.running {
		return nil
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", loopbackAddress, n.settings.Port))
	if err != nil {
		return err
	}

	n.engine.Listener = listener
	n.running = true

	go func() {
		if err := n.engine.Server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			n.running = false
		}
	}()

	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	n.running = false
	return n.engine.Shutdown(ctx)
}

func (n *Node) String() string {
	return fmt.Sprintf("http://%s:%d", loopbackAddress, n.Port())
}
