package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func TestHTTPPolicyFlagsRejectInvalidFerretPolicy(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{name: "allowed scheme", arg: "--policy-http-allowed-schemes=not a scheme", want: "WithAllowedSchemes"},
		{name: "allowed method", arg: "--policy-http-allowed-methods=bad method", want: "WithAllowedMethods"},
		{name: "allowed host", arg: "--policy-http-allowed-hosts=bad host", want: "WithAllowedHosts"},
		{name: "blocked host", arg: "--policy-http-blocked-hosts=bad host", want: "WithBlockedHosts"},
		{name: "default header", arg: `--policy-http-default-headers={"Host":"example.test"}`, want: "WithDefaultHeaders"},
		{name: "blocked header", arg: "--policy-http-blocked-request-headers=bad header", want: "WithBlockedRequestHeaders"},
		{name: "timeout", arg: "--policy-http-timeout=-1s", want: "WithTimeout"},
		{name: "request size", arg: "--policy-http-max-request-size=-1", want: "WithMaxRequestSize"},
		{name: "response size", arg: "--policy-http-max-response-size=-1", want: "WithMaxResponseSize"},
		{name: "response header size", arg: "--policy-http-max-response-header-size=-1", want: "WithMaxResponseHeaderSize"},
		{name: "redirect count", arg: "--policy-http-max-redirects=-1", want: "WithMaxRedirects"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runHTTPPolicyCommand(t, tt.arg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %s error, got %v", tt.want, err)
			}
		})
	}
}

func TestHTTPPolicyFlagsRejectConflictingLimits(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "timeout",
			args: []string{"--policy-http-timeout=1s", "--policy-http-no-timeout"},
			want: "--policy-http-no-timeout cannot be combined with --policy-http-timeout",
		},
		{
			name: "request size",
			args: []string{"--policy-http-max-request-size=1", "--policy-http-unlimited-request-size"},
			want: "--policy-http-unlimited-request-size cannot be combined with --policy-http-max-request-size",
		},
		{
			name: "response size",
			args: []string{"--policy-http-max-response-size=1", "--policy-http-unlimited-response-size"},
			want: "--policy-http-unlimited-response-size cannot be combined with --policy-http-max-response-size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runHTTPPolicyCommand(t, tt.args...)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("expected %q, got %v", tt.want, err)
			}
		})
	}
}

func runHTTPPolicyCommand(t *testing.T, args ...string) error {
	t.Helper()

	command := &cli.Command{
		Name:  "run",
		Flags: httpPolicyFlags(false),
		Action: func(_ context.Context, cmd *cli.Command) error {
			options, err := httpPolicyOptionsFromCommand(cmd)
			if err != nil {
				return err
			}

			client, err := ferrethttp.New(options...)
			if err != nil {
				return err
			}

			if closer, ok := client.(ferrethttp.IdleConnectionCloser); ok {
				closer.CloseIdleConnections()
			}

			return nil
		},
	}

	return command.Run(context.Background(), append([]string{"run"}, args...))
}
