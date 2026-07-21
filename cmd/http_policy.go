package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/runtime"
)

const (
	defaultHTTPPolicyTimeout               = 30 * time.Second
	defaultHTTPPolicyMaxRequestSize  int64 = 16 << 20
	defaultHTTPPolicyMaxResponseSize int64 = 16 << 20
	defaultHTTPPolicyMaxHeaderSize   int64 = 1 << 20
	defaultHTTPPolicyMaxRedirects          = 10
)

func httpPolicyFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "policy-http-allowed-schemes",
			Usage:   "allowed outbound HTTP URL schemes",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_ALLOWED_SCHEMES"),
			Value:   []string{"http", "https"},
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "policy-http-allowed-methods",
			Usage:   "allowed outbound HTTP methods",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_ALLOWED_METHODS"),
			Value:   []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "policy-http-allowed-hosts",
			Usage:   "allowed outbound HTTP hosts (exact host or host:port)",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_ALLOWED_HOSTS"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "policy-http-blocked-hosts",
			Usage:   "blocked outbound HTTP hosts (exact host or host:port)",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_BLOCKED_HOSTS"),
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-http-allow-localhost",
			Usage:   "allow outbound HTTP access to localhost and loopback addresses",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_ALLOW_LOCALHOST"),
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-http-allow-private-networks",
			Usage:   "allow outbound HTTP access to private network addresses",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_ALLOW_PRIVATE_NETWORKS"),
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-http-allow-link-local",
			Usage:   "allow outbound HTTP access to link-local addresses",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_ALLOW_LINK_LOCAL"),
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "policy-http-default-headers",
			Usage:   "default outbound HTTP headers as a JSON object",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_DEFAULT_HEADERS"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "policy-http-blocked-request-headers",
			Usage:   "blocked outbound HTTP request header names",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_BLOCKED_REQUEST_HEADERS"),
			Hidden:  hidden,
		},
		&cli.DurationFlag{
			Name:    "policy-http-timeout",
			Usage:   "overall outbound HTTP timeout",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_TIMEOUT"),
			Value:   defaultHTTPPolicyTimeout,
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-http-no-timeout",
			Usage:   "disable the overall outbound HTTP timeout",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_NO_TIMEOUT"),
			Hidden:  hidden,
		},
		&cli.Int64Flag{
			Name:    "policy-http-max-request-size",
			Usage:   "maximum outbound HTTP request body size in bytes",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_MAX_REQUEST_SIZE"),
			Value:   defaultHTTPPolicyMaxRequestSize,
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-http-unlimited-request-size",
			Usage:   "disable the outbound HTTP request body size limit",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_UNLIMITED_REQUEST_SIZE"),
			Hidden:  hidden,
		},
		&cli.Int64Flag{
			Name:    "policy-http-max-response-size",
			Usage:   "maximum outbound HTTP response body size in bytes",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_MAX_RESPONSE_SIZE"),
			Value:   defaultHTTPPolicyMaxResponseSize,
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-http-unlimited-response-size",
			Usage:   "disable the outbound HTTP response body size limit",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_UNLIMITED_RESPONSE_SIZE"),
			Hidden:  hidden,
		},
		&cli.Int64Flag{
			Name:    "policy-http-max-response-header-size",
			Usage:   "maximum outbound HTTP response header size in bytes",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_MAX_RESPONSE_HEADER_SIZE"),
			Value:   defaultHTTPPolicyMaxHeaderSize,
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-http-follow-redirects",
			Usage:   "follow outbound HTTP redirects",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_FOLLOW_REDIRECTS"),
			Value:   true,
			Hidden:  hidden,
		},
		&cli.IntFlag{
			Name:    "policy-http-max-redirects",
			Usage:   "maximum number of outbound HTTP redirects",
			Sources: cli.EnvVars("LAB_POLICY_HTTP_MAX_REDIRECTS"),
			Value:   defaultHTTPPolicyMaxRedirects,
			Hidden:  hidden,
		},
	}
}

func httpPolicyFromCommand(cmd *cli.Command) (*runtime.HTTPPolicy, error) {
	if cmd == nil {
		return nil, nil
	}

	if cmd.Bool("policy-http-no-timeout") && cmd.IsSet("policy-http-timeout") {
		return nil, fmt.Errorf("--policy-http-no-timeout cannot be combined with --policy-http-timeout")
	}

	if cmd.Bool("policy-http-unlimited-request-size") && cmd.IsSet("policy-http-max-request-size") {
		return nil, fmt.Errorf("--policy-http-unlimited-request-size cannot be combined with --policy-http-max-request-size")
	}

	if cmd.Bool("policy-http-unlimited-response-size") && cmd.IsSet("policy-http-max-response-size") {
		return nil, fmt.Errorf("--policy-http-unlimited-response-size cannot be combined with --policy-http-max-response-size")
	}

	policy := &runtime.HTTPPolicy{}
	configured := false

	if cmd.IsSet("policy-http-allowed-schemes") {
		policy.AllowedSchemes = cmd.StringSlice("policy-http-allowed-schemes")
		configured = true
	}

	if cmd.IsSet("policy-http-allowed-methods") {
		policy.AllowedMethods = cmd.StringSlice("policy-http-allowed-methods")
		configured = true
	}

	if cmd.IsSet("policy-http-allowed-hosts") {
		policy.AllowedHosts = cmd.StringSlice("policy-http-allowed-hosts")
		configured = true
	}

	if cmd.IsSet("policy-http-blocked-hosts") {
		policy.BlockedHosts = cmd.StringSlice("policy-http-blocked-hosts")
		configured = true
	}

	if cmd.IsSet("policy-http-allow-localhost") {
		value := cmd.Bool("policy-http-allow-localhost")
		policy.AllowLocalhost = &value
		configured = true
	}

	if cmd.IsSet("policy-http-allow-private-networks") {
		value := cmd.Bool("policy-http-allow-private-networks")
		policy.AllowPrivateNetworks = &value
		configured = true
	}

	if cmd.IsSet("policy-http-allow-link-local") {
		value := cmd.Bool("policy-http-allow-link-local")
		policy.AllowLinkLocal = &value
		configured = true
	}

	if cmd.IsSet("policy-http-default-headers") {
		headers := make(map[string]string)
		if err := json.Unmarshal([]byte(cmd.String("policy-http-default-headers")), &headers); err != nil {
			return nil, fmt.Errorf("invalid --policy-http-default-headers: expected a JSON object of string values: %w", err)
		}

		if headers == nil {
			return nil, fmt.Errorf("invalid --policy-http-default-headers: expected a JSON object of string values")
		}

		policy.DefaultHeaders = headers
		configured = true
	}

	if cmd.IsSet("policy-http-blocked-request-headers") {
		policy.BlockedRequestHeaders = cmd.StringSlice("policy-http-blocked-request-headers")
		configured = true
	}

	if cmd.IsSet("policy-http-timeout") {
		value := cmd.Duration("policy-http-timeout")
		policy.Timeout = &value
		configured = true
	}

	if cmd.IsSet("policy-http-no-timeout") {
		value := cmd.Bool("policy-http-no-timeout")
		policy.NoTimeout = &value
		configured = true
	}

	if cmd.IsSet("policy-http-max-request-size") {
		value := cmd.Int64("policy-http-max-request-size")
		policy.MaxRequestSize = &value
		configured = true
	}

	if cmd.IsSet("policy-http-unlimited-request-size") {
		value := cmd.Bool("policy-http-unlimited-request-size")
		policy.UnlimitedRequestSize = &value
		configured = true
	}

	if cmd.IsSet("policy-http-max-response-size") {
		value := cmd.Int64("policy-http-max-response-size")
		policy.MaxResponseSize = &value
		configured = true
	}

	if cmd.IsSet("policy-http-unlimited-response-size") {
		value := cmd.Bool("policy-http-unlimited-response-size")
		policy.UnlimitedResponseSize = &value
		configured = true
	}

	if cmd.IsSet("policy-http-max-response-header-size") {
		value := cmd.Int64("policy-http-max-response-header-size")
		policy.MaxResponseHeaderSize = &value
		configured = true
	}

	if cmd.IsSet("policy-http-follow-redirects") {
		value := cmd.Bool("policy-http-follow-redirects")
		policy.FollowRedirects = &value
		configured = true
	}

	if cmd.IsSet("policy-http-max-redirects") {
		value := cmd.Int("policy-http-max-redirects")
		policy.MaxRedirects = &value
		configured = true
	}

	if !configured {
		return nil, nil
	}

	return policy, nil
}
