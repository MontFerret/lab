package staticserver

import (
	"context"
	"net"
	"testing"
)

func TestResolveSettingsDefaults(t *testing.T) {
	settings, err := ResolveSettings(Settings{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if settings.BindHost != defaultHost {
		t.Fatalf("expected default bind host %q, got %q", defaultHost, settings.BindHost)
	}

	if settings.AdvertiseHost != defaultHost {
		t.Fatalf("expected default advertised host %q, got %q", defaultHost, settings.AdvertiseHost)
	}
}

func TestResolveSettingsUsesWildcardBindForAdvertisedHost(t *testing.T) {
	settings, err := ResolveSettings(Settings{AdvertiseHost: "example.test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if settings.BindHost != "0.0.0.0" {
		t.Fatalf("expected wildcard IPv4 bind host, got %q", settings.BindHost)
	}

	if settings.AdvertiseHost != "example.test" {
		t.Fatalf("expected advertised host to be preserved, got %q", settings.AdvertiseHost)
	}
}

func TestResolveSettingsUsesWildcardIPv6BindForIPv6AdvertisedHost(t *testing.T) {
	settings, err := ResolveSettings(Settings{AdvertiseHost: "::1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if settings.BindHost != "::" {
		t.Fatalf("expected wildcard IPv6 bind host, got %q", settings.BindHost)
	}
}

func TestResolveSettingsRejectsHostsWithPorts(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
	}{
		{
			name:     "bind host includes port",
			settings: Settings{BindHost: "127.0.0.1:8080"},
		},
		{
			name:     "advertised host includes port",
			settings: Settings{AdvertiseHost: "example.test:8080"},
		},
		{
			name:     "ipv6 host includes port",
			settings: Settings{AdvertiseHost: "[::1]:8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ResolveSettings(tt.settings); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestManagerEndpointsUseAdvertisedHost(t *testing.T) {
	manager, err := NewManager(Settings{
		BindHost:      "0.0.0.0",
		AdvertiseHost: "example.test",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	dir := t.TempDir()
	if err := manager.Bind(ServeEntry{Alias: "app", Path: dir, Port: 8080}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := manager.Endpoints()["app"]; got != "http://example.test:8080" {
		t.Fatalf("expected advertised endpoint, got %q", got)
	}
}

func TestManagerEndpointsFormatIPv6Hosts(t *testing.T) {
	manager, err := NewManager(Settings{AdvertiseHost: "::1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	dir := t.TempDir()
	if err := manager.Bind(ServeEntry{Alias: "app", Path: dir, Port: 8080}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := manager.Endpoints()["app"]; got != "http://[::1]:8080" {
		t.Fatalf("expected IPv6 endpoint formatting, got %q", got)
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
		Dir:           t.TempDir(),
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

	addr, ok := node.engine.Listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("expected TCP listener address, got %T", node.engine.Listener.Addr())
	}

	if !addr.IP.IsLoopback() {
		t.Fatalf("expected loopback bind host, got %s", addr.IP.String())
	}
}
