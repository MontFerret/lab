package staticserver

import (
	"context"
	"net/http"

	"github.com/MontFerret/lab/v2/pkg/localserver"
)

type (
	StaticEndpoints map[string]string

	Manager struct {
		inner *localserver.Manager
	}
)

func NewManager(settings Settings) (*Manager, error) {
	inner, err := localserver.NewManager(localserver.ManagerOptions{
		Settings: settings,
		HandlerFactory: func(entry localserver.Entry) (http.Handler, error) {
			return newStaticHandler(entry.Path, ""), nil
		},
		StartErrorLabel: "failed to start static file server",
		StopErrorLabel:  "failed to stop static file server",
	})
	if err != nil {
		return nil, err
	}

	return &Manager{inner: inner}, nil
}

func (m *Manager) IsRunning() bool {
	return m.inner.IsRunning()
}

func (m *Manager) Bind(entry ServeEntry) error {
	return m.inner.Bind(entry)
}

func (m *Manager) Endpoints() StaticEndpoints {
	source := m.inner.Endpoints()
	endpoints := make(StaticEndpoints, len(source))

	for name, endpoint := range source {
		endpoints[name] = endpoint
	}

	return endpoints
}

func (m *Manager) Start(ctx context.Context) error {
	return m.inner.Start(ctx)
}

func (m *Manager) Stop(ctx context.Context) error {
	return m.inner.Stop(ctx)
}
