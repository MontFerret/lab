package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"

	"github.com/MontFerret/lab/v2/pkg/runtime"
)

func TestHTTPPolicyFlagsRejectInvalidFerretPolicy(t *testing.T) {
	tests := []struct {
		name      string
		arg       string
		wantField string
		wantValue string
	}{
		{name: "allowed scheme", arg: "--policy-http-allowed-schemes=not a scheme", wantField: "allowed schemes", wantValue: "not a scheme"},
		{name: "allowed method", arg: "--policy-http-allowed-methods=bad method", wantField: "allowed methods", wantValue: "bad method"},
		{name: "allowed host", arg: "--policy-http-allowed-hosts=bad host", wantField: "allowed hosts", wantValue: "bad host"},
		{name: "blocked host", arg: "--policy-http-blocked-hosts=bad host", wantField: "blocked hosts", wantValue: "bad host"},
		{name: "default header", arg: `--policy-http-default-headers={"Host":"example.test"}`, wantField: "default headers", wantValue: "Host"},
		{name: "blocked header", arg: "--policy-http-blocked-request-headers=bad header", wantField: "blocked request headers", wantValue: "bad header"},
		{name: "timeout", arg: "--policy-http-timeout=-1s", wantField: "timeout", wantValue: "-1s"},
		{name: "request size", arg: "--policy-http-max-request-size=-1", wantField: "max request size", wantValue: "-1"},
		{name: "response size", arg: "--policy-http-max-response-size=-1", wantField: "max response size", wantValue: "-1"},
		{name: "response header size", arg: "--policy-http-max-response-header-size=-1", wantField: "max response header size", wantValue: "-1"},
		{name: "redirect count", arg: "--policy-http-max-redirects=-1", wantField: "max redirects", wantValue: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runHTTPPolicyCommand(t, tt.arg)
			if !errors.Is(err, ferrethttp.ErrInvalidPolicyConfiguration) {
				t.Fatalf("expected invalid HTTP policy configuration error, got %v", err)
			}

			if !strings.Contains(err.Error(), tt.wantField) || !strings.Contains(err.Error(), "value="+tt.wantValue) {
				t.Fatalf("expected %s diagnostic for %q, got %v", tt.wantField, tt.wantValue, err)
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
			policy, err := httpPolicyFromCommand(cmd)
			if err != nil {
				return err
			}

			rt, err := runtime.New(runtime.Options{HTTPPolicy: policy})
			if err != nil {
				return err
			}

			return rt.Close()
		},
	}

	return command.Run(context.Background(), append([]string{"run"}, args...))
}
