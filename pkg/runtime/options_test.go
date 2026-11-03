package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestRuntimeRejectsInvalidTargets(t *testing.T) {
	for _, target := range []string{
		"wire", " W i R e ", "unknown", "unknown://example", "tcp://127.0.0.1:42", "%",
		"http:", "http:///run", "https://", "http://host:", "http://host:0", "http://host:65536", "http://host:no", "http://::1", "http://host:80:90",
		"wire+tcp:", "wire+tcp:///", "wire+tcp://127.0.0.1", "wire+tcp://127.0.0.1:",
		"wire+tcp://127.0.0.1:0", "wire+tcp://127.0.0.1:65536", "wire+tcp://127.0.0.1:-1", "wire+tcp://127.0.0.1:no",
		"wire+tcp://localhost:42", "wire+tcp://127.0.0.2:42", "wire+tcp://[::1]:42", "wire+tcp://[::1:42",
		"wire+tcp://user:password@127.0.0.1:42", "wire+tcp://127.0.0.1:42/", "wire+tcp://127.0.0.1:42/path",
		"wire+tcp://127.0.0.1:42?", "wire+tcp://127.0.0.1:42?x=1", "wire+tcp://127.0.0.1:42#", "wire+tcp://127.0.0.1:42#x",
		" wire+tcp://127.0.0.1:42", "wire+tcp://127.0.0.1:42 ",
	} {
		t.Run(target, func(t *testing.T) {
			opts := Options{Target: target}
			if err := opts.Validate(); err == nil {
				t.Fatal("invalid target passed preflight")
			}

			rt, err := New(t.Context(), opts)
			if err == nil || rt != nil {
				t.Fatalf("invalid target constructed %T, %v", rt, err)
			}
		})
	}
}

func TestWireResolvedConnectionConfiguration(t *testing.T) {
	for _, tc := range []struct {
		target   string
		options  map[string]any
		endpoint string
		timeout  time.Duration
	}{
		{"wire+tcp://127.0.0.1:1", nil, "tcp://127.0.0.1:1", 5 * time.Second},
		{"wire+tcp://127.0.0.1:65535", nil, "tcp://127.0.0.1:65535", 5 * time.Second},
		{"wire+tcp://127.0.0.1:00042", map[string]any{"connect-timeout": "10s"}, "tcp://127.0.0.1:42", 10 * time.Second},
	} {
		config, err := (Options{Target: tc.target, AdapterOptions: tc.options}).resolve()
		if err != nil || config.endpoint != tc.endpoint || config.connectTimeout != tc.timeout {
			t.Fatalf("connection configuration: %#v, %v", config, err)
		}
	}
}

func TestRuntimeOptionValidation(t *testing.T) {
	for _, target := range []string{"", "builtin", "http://worker.test/api", "https://worker.test/api", "wire+tcp://127.0.0.1:42"} {
		for _, options := range []map[string]any{
			{"value": "shared"}, {"flags": []any{}},
			{"connect-timeout": nil}, {"connect-timeout": 5}, {"connect-timeout": true},
			{"connect-timeout": "0s"}, {"connect-timeout": "-1s"}, {"connect-timeout": "invalid"}, {"connect-timeout": "999999999999h"},
			{"headers": "secret"}, {"headers": nil}, {"headers": map[string]any{"Authorization": 42}},
			{"headers": map[string]any{"Authorization": "secret\nvalue"}},
			{"cookies": []any{}}, {"cookies": map[string]any{"session": false}},
			{"path": 42}, {"path": "%"}, {"path": "http://"}, {"path": "ftp://host"},
		} {
			opts := Options{Target: target, AdapterOptions: options}
			err := opts.Validate()
			if err == nil || !strings.Contains(err.Error(), "runtime option") || strings.Contains(err.Error(), "secret") {
				t.Fatalf("option validation for %q: %v", target, err)
			}
		}
	}

	for _, opts := range []Options{
		{Target: "wire+tcp://127.0.0.1:42", AdapterOptions: map[string]any{"headers": map[string]any{}}},
		{Target: "wire+tcp://127.0.0.1:42", FSPolicy: &FileSystemPolicy{ReadOnly: pointerTo(false)}},
		{Target: "wire+tcp://127.0.0.1:42", HTTPPolicy: &HTTPPolicy{NoTimeout: pointerTo(false)}},
	} {
		if err := opts.Validate(); err == nil {
			t.Fatal("unsupported Wire configuration passed preflight")
		}
	}
}
