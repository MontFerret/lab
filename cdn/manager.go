package cdn

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
)

type Manager struct {
	server  *http.Server
	nodes   []*Node
	running bool
}

func New() *Manager {
	return &Manager{
		nodes: make([]*Node, 0, 10),
	}
}

func (m *Manager) IsRunning() bool {
	return m.running
}

func (m *Manager) Add(node *Node) error {
	m.nodes = append(m.nodes, node)

	return nil
}

func (m *Manager) Endpoints() map[string]string {
	endpoints := make(map[string]string)

	for _, n := range m.nodes {
		endpoints[n.Name()] = fmt.Sprintf("http://127.0.0.1:%d", n.Port())
	}

	return endpoints
}

func (m *Manager) Start(ctx context.Context) error {
	var s sync.RWMutex
	var w sync.WaitGroup

	failed := make(map[int]error)

	for _, n := range m.nodes {
		if n.IsRunning() {
			continue
		}

		w.Add(1)

		go func() {
			w.Done()

			err := n.Start(ctx)

			if err != nil {
				s.Lock()

				failed[n.ID()] = err

				s.Unlock()
			}
		}()
	}

	w.Wait()

	if len(failed) == 0 {
		return nil
	}

	// stop running nodes
	for _, n := range m.nodes {
		_, exists := failed[n.ID()]

		if !exists {
			n.Stop(ctx)
		}
	}

	return errors.Errorf("failed to start static server(s): %s", m.joinErrors(failed))
}

func (m *Manager) Stop(ctx context.Context) error {
	failed := make(map[int]error)

	for _, n := range m.nodes {
		if !n.IsRunning() {
			continue
		}

		if err := n.Stop(ctx); err != nil {
			failed[n.ID()] = err
		}
	}

	if len(failed) == 0 {
		return nil
	}

	return errors.Errorf("failed to stop static server(s): %s", m.joinErrors(failed))
}

func (m *Manager) joinErrors(failed map[int]error) string {
	var buf bytes.Buffer
	var i int
	last := len(failed) - 1

	for _, e := range failed {
		buf.WriteString(e.Error())

		if i < last {
			buf.WriteString(",")
		}

		i++
	}

	return buf.String()
}
