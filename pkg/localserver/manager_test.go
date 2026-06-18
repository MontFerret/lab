package localserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestManagerServesHandlerAndTracksEndpoints(t *testing.T) {
	manager, err := NewManager(ManagerOptions{
		Settings: Settings{},
		HandlerFactory: func(entry Entry) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(entry.Alias))
			}), nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := manager.Bind(Entry{Alias: "api", Path: "unused"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Stop(ctx)
	}()

	endpoint := manager.Endpoints()["api"]
	resp, err := http.Get(endpoint)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(body) != "api" {
		t.Fatalf("expected body api, got %q", string(body))
	}
}

func TestManagerEndpointsUseAdvertisedHost(t *testing.T) {
	manager, err := NewManager(ManagerOptions{
		Settings: Settings{
			BindHost:      "0.0.0.0",
			AdvertiseHost: "example.test",
		},
		HandlerFactory: func(_ Entry) (http.Handler, error) {
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := manager.Bind(Entry{Alias: "app", Port: 8080}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := manager.Endpoints()["app"]; got != "http://example.test:8080" {
		t.Fatalf("expected advertised endpoint, got %q", got)
	}
}

func TestNodeStartUsesBindHost(t *testing.T) {
	port, err := GetFreePort("127.0.0.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	node, err := NewNode(NodeSettings{
		Name:          "app",
		Port:          port,
		Handler:       http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		BindHost:      "127.0.0.1",
		AdvertiseHost: "example.test",
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

	addr, ok := node.ListenerAddr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("expected TCP listener address, got %T", node.ListenerAddr())
	}

	if !addr.IP.IsLoopback() {
		t.Fatalf("expected loopback bind host, got %s", addr.IP.String())
	}
}
