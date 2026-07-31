package runtime

import (
	"net/url"
	"testing"
)

func TestParseProtocol(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  Protocol
	}{
		{name: "CLI", value: "cli", want: ProtocolCLI},
		{name: "normalized CLI", value: "  CLI\t", want: ProtocolCLI},
		{name: "CLI direct", value: "cli-direct", want: ProtocolCLIDirect},
		{name: "normalized CLI direct", value: "\nCLI-DIRECT ", want: ProtocolCLIDirect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProtocol(tt.value)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestParseProtocolRejectsUnknownValue(t *testing.T) {
	protocol, err := ParseProtocol("custom")
	if protocol != ProtocolUnknown {
		t.Fatalf("expected unknown protocol, got %q", protocol)
	}
	if err == nil || err.Error() != `unknown runtime protocol "custom"; expected one of: cli, cli-direct` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseProtocolFromURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   Protocol
	}{
		{name: "missing query defaults to fallback", rawURL: "bin://./bin/runtime", want: ProtocolCLIDirect},
		{name: "empty query defaults to fallback", rawURL: "bin://./bin/runtime?protocol=", want: ProtocolCLIDirect},
		{name: "explicit CLI", rawURL: "bin://./bin/runtime?protocol=cli", want: ProtocolCLI},
		{name: "explicit CLI direct", rawURL: "bin://./bin/runtime?protocol=cli-direct", want: ProtocolCLIDirect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("parse URL: %v", err)
			}

			got, err := ParseProtocolFrom(u, ProtocolCLIDirect)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNewParsesBinaryProtocolFromRuntimeURL(t *testing.T) {
	tests := []struct {
		name     string
		runtime  string
		wantPath string
		want     Protocol
	}{
		{
			name:     "opaque path defaults to CLI direct",
			runtime:  "bin:./bin/runtime",
			wantPath: "./bin/runtime",
			want:     ProtocolCLIDirect,
		},
		{
			name:     "opaque path with CLI direct",
			runtime:  "bin:./bin/runtime?protocol=cli-direct",
			wantPath: "./bin/runtime",
			want:     ProtocolCLIDirect,
		},
		{
			name:     "URL path with CLI",
			runtime:  "bin://./bin/runtime?protocol=cli",
			wantPath: "./bin/runtime",
			want:     ProtocolCLI,
		},
		{
			name:     "URL path with CLI direct",
			runtime:  "bin://./bin/runtime?protocol=cli-direct",
			wantPath: "./bin/runtime",
			want:     ProtocolCLIDirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := New(Options{Type: tt.runtime})
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			binary, ok := rt.(*Binary)
			if !ok {
				t.Fatalf("expected binary runtime, got %T", rt)
			}
			if binary.path != tt.wantPath {
				t.Fatalf("expected path %q, got %q", tt.wantPath, binary.path)
			}
			if binary.protocol != tt.want {
				t.Fatalf("expected protocol %q, got %q", tt.want, binary.protocol)
			}
		})
	}
}

func TestNewRejectsUnknownBinaryProtocol(t *testing.T) {
	_, err := New(Options{Type: "bin://./bin/runtime?protocol=custom"})
	if err == nil || err.Error() != `failed to parse binary runtime protocol: unknown runtime protocol "custom"; expected one of: cli, cli-direct` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewBinaryRejectsUnknownProtocol(t *testing.T) {
	_, err := NewBinary(BinaryOptions{
		Path:     "./bin/runtime",
		Protocol: Protocol("custom"),
	})
	if err == nil || err.Error() != `unknown runtime protocol "custom"; expected one of: cli, cli-direct` {
		t.Fatalf("unexpected error: %v", err)
	}
}
