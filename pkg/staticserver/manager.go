package staticserver

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
)

const loopbackAddress = "127.0.0.1"

type (
	StaticEndpoints map[string]string

	Manager struct {
		nodes   []*Node
		running bool
	}
)

func NewManager() *Manager {
	return &Manager{
		nodes: make([]*Node, 0, 10),
	}
}

func (m *Manager) IsRunning() bool {
	return m.running
}

func (m *Manager) Bind(entry ServeEntry) error {
	port := entry.Port
	if port == 0 {
		assigned, err := GetFreePort()
		if err != nil {
			return err
		}

		port = assigned
	}

	node, err := NewNode(NodeSettings{
		Name:   entry.Alias,
		Port:   port,
		Dir:    entry.Path,
		Prefix: "",
	})
	if err != nil {
		return err
	}

	m.nodes = append(m.nodes, node)

	return nil
}

func (m *Manager) Endpoints() StaticEndpoints {
	endpoints := make(StaticEndpoints, len(m.nodes))

	for _, node := range m.nodes {
		endpoints[node.Name()] = fmt.Sprintf("http://%s:%d", loopbackAddress, node.Port())
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

	return errors.Errorf("failed to start static file server: %s", m.joinErrors(failed))
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

	return errors.Errorf("failed to stop static file server: %s", m.joinErrors(failed))
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
