package cdn

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"math/rand"
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

func NewNode(settings NodeSettings) *Node {
	e := echo.New()
	e.Debug = false
	e.HideBanner = true
	e.Use(middleware.CORS())

	prefix := "/"

	if settings.Prefix != "" {
		prefix = settings.Prefix
	}

	e.Static(prefix, settings.Dir)

	return &Node{
		id:       rand.Int(),
		settings: settings,
		engine:   e,
	}
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
	n.running = true
	err := n.engine.Start(fmt.Sprintf("0.0.0.0:%d", n.settings.Port))

	if err != nil {
		n.running = false
	}

	return err
}

func (n *Node) Stop(ctx context.Context) error {
	n.running = false
	return n.engine.Shutdown(ctx)
}

func (n *Node) String() string {
	return fmt.Sprintf("http://127.0.0.1:%d", n.Port())
}
