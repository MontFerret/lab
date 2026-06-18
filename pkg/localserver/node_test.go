package localserver

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestNodeIsRunningReflectsLifecycle(t *testing.T) {
	port, err := GetFreePort("127.0.0.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	node, err := NewNode(NodeSettings{
		Name:     "test",
		Port:     port,
		Handler:  http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		BindHost: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if node.IsRunning() {
		t.Fatal("expected node to not be running before Start")
	}

	if err := node.Start(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !node.IsRunning() {
		t.Fatal("expected node to be running after Start")
	}

	if err := node.Stop(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if node.IsRunning() {
		t.Fatal("expected node to not be running after Stop")
	}
}

func TestNodeIsRunningIsSafeForConcurrentAccess(t *testing.T) {
	port, err := GetFreePort("127.0.0.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	node, err := NewNode(NodeSettings{
		Name:     "race",
		Port:     port,
		Handler:  http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		BindHost: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := node.Start(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer func() {
		_ = node.Stop(context.Background())
	}()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = node.IsRunning()
		}()
	}
	wg.Wait()
}

func TestNodeServeErrIsNilOnHealthyNode(t *testing.T) {
	port, err := GetFreePort("127.0.0.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	node, err := NewNode(NodeSettings{
		Name:     "healthy",
		Port:     port,
		Handler:  http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		BindHost: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := node.Start(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer func() {
		_ = node.Stop(context.Background())
	}()

	if got := node.ServeErr(); got != nil {
		t.Fatalf("expected no serve error, got %v", got)
	}
}

func TestNodeStartFailsWithInvalidBindHost(t *testing.T) {
	node, err := NewNode(NodeSettings{
		Name:     "badhost",
		Port:     0,
		Handler:  http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		BindHost: "192.0.2.1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = node.Start(context.Background())
	if err == nil {
		defer func() { _ = node.Stop(context.Background()) }()
		t.Fatal("expected error when binding to unreachable host")
	}

	if node.IsRunning() {
		t.Fatal("expected node to not be running after failed Start")
	}
}

func TestNodeIDsAreUnique(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	ids := make(map[int]bool)

	for i := 0; i < 100; i++ {
		node, err := NewNode(NodeSettings{
			Name:     "unique",
			Port:     0,
			Handler:  handler,
			BindHost: "127.0.0.1",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ids[node.ID()] {
			t.Fatalf("duplicate node ID %d", node.ID())
		}
		ids[node.ID()] = true
	}
}

func TestManagerIsRunningReflectsLifecycle(t *testing.T) {
	manager, err := NewManager(ManagerOptions{
		Settings: Settings{},
		HandlerFactory: func(_ Entry) (http.Handler, error) {
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if manager.IsRunning() {
		t.Fatal("expected manager to not be running before Start")
	}

	if err := manager.Bind(Entry{Alias: "svc"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !manager.IsRunning() {
		t.Fatal("expected manager to be running after Start")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := manager.Stop(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if manager.IsRunning() {
		t.Fatal("expected manager to not be running after Stop")
	}
}
