package staticserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/MontFerret/lab/v2/pkg/localserver"
)

type (
	NodeSettings struct {
		Name          string
		Port          int
		Dir           string
		Prefix        string
		BindHost      string
		AdvertiseHost string
	}

	Node = localserver.Node
)

func NewNode(settings NodeSettings) (*Node, error) {
	handler := newStaticHandler(settings.Dir, settings.Prefix)

	node, err := localserver.NewNode(localserver.NodeSettings{
		Name:          settings.Name,
		Port:          settings.Port,
		Handler:       handler,
		BindHost:      settings.BindHost,
		AdvertiseHost: settings.AdvertiseHost,
	})
	if err != nil {
		return nil, err
	}

	return node, nil
}

func newStaticHandler(dir string, configuredPrefix string) http.Handler {
	engine := echo.New()
	engine.Debug = false
	engine.HideBanner = true
	engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodOptions},
		AllowHeaders: []string{"*"},
	}))

	prefix := "/"
	if configuredPrefix != "" {
		prefix = configuredPrefix
	}

	engine.Static(prefix, dir)

	return engine
}
