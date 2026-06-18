package localserver

import (
	"bytes"
	"context"
	"net/http"

	"github.com/pkg/errors"
)

type (
	Endpoints map[string]string

	HandlerFactory func(entry Entry) (http.Handler, error)

	ManagerOptions struct {
		Settings        Settings
		HandlerFactory  HandlerFactory
		StartErrorLabel string
		StopErrorLabel  string
	}

	Manager struct {
		settings        Settings
		handlerFactory  HandlerFactory
		startErrorLabel string
		stopErrorLabel  string
		nodes           []*Node
		running         bool
	}
)

func NewManager(opts ManagerOptions) (*Manager, error) {
	if opts.HandlerFactory == nil {
		return nil, errors.New("local server handler factory is required")
	}

	resolved, err := ResolveSettings(opts.Settings)
	if err != nil {
		return nil, err
	}

	return &Manager{
		settings:        resolved,
		handlerFactory:  opts.HandlerFactory,
		startErrorLabel: labelOrDefault(opts.StartErrorLabel, "failed to start local server"),
		stopErrorLabel:  labelOrDefault(opts.StopErrorLabel, "failed to stop local server"),
		nodes:           make([]*Node, 0, 10),
	}, nil
}

func (m *Manager) IsRunning() bool {
	return m.running
}

func (m *Manager) Bind(entry Entry) error {
	port := entry.Port
	if port == 0 {
		assigned, err := GetFreePort(m.settings.BindHost)
		if err != nil {
			return err
		}

		port = assigned
	}

	entry.Port = port
	handler, err := m.handlerFactory(entry)
	if err != nil {
		return err
	}

	node, err := NewNode(NodeSettings{
		Name:          entry.Alias,
		Port:          port,
		Handler:       handler,
		BindHost:      m.settings.BindHost,
		AdvertiseHost: m.settings.AdvertiseHost,
	})
	if err != nil {
		return err
	}

	m.nodes = append(m.nodes, node)

	return nil
}

func (m *Manager) Endpoints() Endpoints {
	endpoints := make(Endpoints, len(m.nodes))

	for _, node := range m.nodes {
		endpoints[node.Name()] = node.String()
	}

	return endpoints
}

func (m *Manager) Start(ctx context.Context) error {
	failed := make(map[int]error)

	for _, node := range m.nodes {
		if node.IsRunning() {
			continue
		}

		if err := node.Start(ctx); err != nil {
			failed[node.ID()] = err
		}
	}

	if len(failed) == 0 {
		m.running = len(m.nodes) > 0
		return nil
	}

	for _, node := range m.nodes {
		if _, exists := failed[node.ID()]; !exists {
			_ = node.Stop(ctx)
		}
	}

	return errors.Errorf("%s: %s", m.startErrorLabel, m.joinErrors(failed))
}

func (m *Manager) Stop(ctx context.Context) error {
	failed := make(map[int]error)

	for _, node := range m.nodes {
		if !node.IsRunning() {
			continue
		}

		if err := node.Stop(ctx); err != nil {
			failed[node.ID()] = err
		}
	}

	m.running = false

	if len(failed) == 0 {
		return nil
	}

	return errors.Errorf("%s: %s", m.stopErrorLabel, m.joinErrors(failed))
}

func (m *Manager) joinErrors(failed map[int]error) string {
	var buf bytes.Buffer
	var i int
	last := len(failed) - 1

	for _, err := range failed {
		buf.WriteString(err.Error())

		if i < last {
			buf.WriteString(",")
		}

		i++
	}

	return buf.String()
}
