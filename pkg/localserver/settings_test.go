package localserver

import "testing"

func TestResolveSettingsDefaults(t *testing.T) {
	settings, err := ResolveSettings(Settings{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if settings.BindHost != DefaultHost {
		t.Fatalf("expected default bind host %q, got %q", DefaultHost, settings.BindHost)
	}

	if settings.AdvertiseHost != DefaultHost {
		t.Fatalf("expected default advertised host %q, got %q", DefaultHost, settings.AdvertiseHost)
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
