package runtime

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

// HTTPPolicy describes outbound HTTP policy overrides shared by built-in and
// binary Ferret runtimes. Nil fields leave the runtime's own setting unchanged.
type HTTPPolicy struct {
	AllowedSchemes        []string
	AllowedMethods        []string
	AllowedHosts          []string
	BlockedHosts          []string
	AllowLocalhost        *bool
	AllowPrivateNetworks  *bool
	AllowLinkLocal        *bool
	DefaultHeaders        map[string]string
	BlockedRequestHeaders []string
	Timeout               *time.Duration
	NoTimeout             *bool
	MaxRequestSize        *int64
	UnlimitedRequestSize  *bool
	MaxResponseSize       *int64
	UnlimitedResponseSize *bool
	MaxResponseHeaderSize *int64
	FollowRedirects       *bool
	MaxRedirects          *int
}

func (policy *HTTPPolicy) hasSettings() bool {
	return policy != nil && (policy.AllowedSchemes != nil ||
		policy.AllowedMethods != nil ||
		policy.AllowedHosts != nil ||
		policy.BlockedHosts != nil ||
		policy.AllowLocalhost != nil ||
		policy.AllowPrivateNetworks != nil ||
		policy.AllowLinkLocal != nil ||
		policy.DefaultHeaders != nil ||
		policy.BlockedRequestHeaders != nil ||
		policy.Timeout != nil ||
		policy.NoTimeout != nil ||
		policy.MaxRequestSize != nil ||
		policy.UnlimitedRequestSize != nil ||
		policy.MaxResponseSize != nil ||
		policy.UnlimitedResponseSize != nil ||
		policy.MaxResponseHeaderSize != nil ||
		policy.FollowRedirects != nil ||
		policy.MaxRedirects != nil)
}

func (policy *HTTPPolicy) validate() error {
	_, err := policy.validatedFerretOptions()

	return err
}

func (policy *HTTPPolicy) validatedFerretOptions() ([]ferrethttp.PolicyOption, error) {
	if policy == nil {
		return nil, nil
	}

	if policy.NoTimeout != nil && *policy.NoTimeout && policy.Timeout != nil {
		return nil, fmt.Errorf("--policy-http-no-timeout cannot be combined with --policy-http-timeout")
	}

	if policy.UnlimitedRequestSize != nil && *policy.UnlimitedRequestSize && policy.MaxRequestSize != nil {
		return nil, fmt.Errorf("--policy-http-unlimited-request-size cannot be combined with --policy-http-max-request-size")
	}

	if policy.UnlimitedResponseSize != nil && *policy.UnlimitedResponseSize && policy.MaxResponseSize != nil {
		return nil, fmt.Errorf("--policy-http-unlimited-response-size cannot be combined with --policy-http-max-response-size")
	}

	var options []ferrethttp.PolicyOption
	if policy.AllowedSchemes != nil {
		options = append(options, ferrethttp.WithAllowedSchemes(policy.AllowedSchemes...))
	}

	if policy.AllowedMethods != nil {
		options = append(options, ferrethttp.WithAllowedMethods(policy.AllowedMethods...))
	}

	if policy.AllowedHosts != nil {
		options = append(options, ferrethttp.WithAllowedHosts(policy.AllowedHosts...))
	}

	if policy.BlockedHosts != nil {
		options = append(options, ferrethttp.WithBlockedHosts(policy.BlockedHosts...))
	}

	if policy.AllowLocalhost != nil {
		options = append(options, ferrethttp.WithAllowLocalhost(*policy.AllowLocalhost))
	}

	if policy.AllowPrivateNetworks != nil {
		options = append(options, ferrethttp.WithAllowPrivateNetworks(*policy.AllowPrivateNetworks))
	}

	if policy.AllowLinkLocal != nil {
		options = append(options, ferrethttp.WithAllowLinkLocal(*policy.AllowLinkLocal))
	}

	if policy.DefaultHeaders != nil {
		options = append(options, ferrethttp.WithDefaultHeaders(policy.DefaultHeaders))
	}

	if policy.BlockedRequestHeaders != nil {
		options = append(options, ferrethttp.WithBlockedRequestHeaders(policy.BlockedRequestHeaders...))
	}

	if policy.NoTimeout != nil && *policy.NoTimeout {
		options = append(options, ferrethttp.WithNoTimeout())
	} else if policy.Timeout != nil {
		options = append(options, ferrethttp.WithTimeout(*policy.Timeout))
	}

	if policy.UnlimitedRequestSize != nil && *policy.UnlimitedRequestSize {
		options = append(options, ferrethttp.WithUnlimitedRequestSize())
	} else if policy.MaxRequestSize != nil {
		options = append(options, ferrethttp.WithMaxRequestSize(*policy.MaxRequestSize))
	}

	if policy.UnlimitedResponseSize != nil && *policy.UnlimitedResponseSize {
		options = append(options, ferrethttp.WithUnlimitedResponseSize())
	} else if policy.MaxResponseSize != nil {
		options = append(options, ferrethttp.WithMaxResponseSize(*policy.MaxResponseSize))
	}

	if policy.MaxResponseHeaderSize != nil {
		options = append(options, ferrethttp.WithMaxResponseHeaderSize(*policy.MaxResponseHeaderSize))
	}

	if policy.FollowRedirects != nil {
		options = append(options, ferrethttp.WithFollowRedirects(*policy.FollowRedirects))
	}

	if policy.MaxRedirects != nil {
		options = append(options, ferrethttp.WithMaxRedirects(*policy.MaxRedirects))
	}

	if _, err := ferrethttp.NewPolicy(options...); err != nil {
		return nil, err
	}

	return options, nil
}

func (policy *HTTPPolicy) ferretCLIArgs() ([]string, error) {
	if policy == nil {
		return nil, nil
	}

	args := make([]string, 0, 18)
	args = appendStringSliceFlag(args, "policy-http-allowed-schemes", policy.AllowedSchemes)
	args = appendStringSliceFlag(args, "policy-http-allowed-methods", policy.AllowedMethods)
	args = appendStringSliceFlag(args, "policy-http-allowed-hosts", policy.AllowedHosts)
	args = appendStringSliceFlag(args, "policy-http-blocked-hosts", policy.BlockedHosts)
	args = appendBoolFlag(args, "policy-http-allow-localhost", policy.AllowLocalhost)
	args = appendBoolFlag(args, "policy-http-allow-private-networks", policy.AllowPrivateNetworks)
	args = appendBoolFlag(args, "policy-http-allow-link-local", policy.AllowLinkLocal)

	if policy.DefaultHeaders != nil {
		data, err := json.Marshal(policy.DefaultHeaders)
		if err != nil {
			return nil, fmt.Errorf("serialize HTTP policy default headers: %w", err)
		}

		args = append(args, "--policy-http-default-headers="+string(data))
	}

	args = appendStringSliceFlag(args, "policy-http-blocked-request-headers", policy.BlockedRequestHeaders)
	if policy.Timeout != nil {
		args = append(args, "--policy-http-timeout="+policy.Timeout.String())
	}

	args = appendBoolFlag(args, "policy-http-no-timeout", policy.NoTimeout)
	if policy.MaxRequestSize != nil {
		args = append(args, "--policy-http-max-request-size="+strconv.FormatInt(*policy.MaxRequestSize, 10))
	}

	args = appendBoolFlag(args, "policy-http-unlimited-request-size", policy.UnlimitedRequestSize)
	if policy.MaxResponseSize != nil {
		args = append(args, "--policy-http-max-response-size="+strconv.FormatInt(*policy.MaxResponseSize, 10))
	}

	args = appendBoolFlag(args, "policy-http-unlimited-response-size", policy.UnlimitedResponseSize)
	if policy.MaxResponseHeaderSize != nil {
		args = append(args, "--policy-http-max-response-header-size="+strconv.FormatInt(*policy.MaxResponseHeaderSize, 10))
	}

	args = appendBoolFlag(args, "policy-http-follow-redirects", policy.FollowRedirects)
	if policy.MaxRedirects != nil {
		args = append(args, "--policy-http-max-redirects="+strconv.Itoa(*policy.MaxRedirects))
	}

	return args, nil
}

func (policy *HTTPPolicy) conflictingRawFlags() map[string]struct{} {
	flags := make(map[string]struct{}, 18)
	if policy == nil {
		return flags
	}

	addManagedSliceFlag(flags, "--policy-http-allowed-schemes", policy.AllowedSchemes)
	addManagedSliceFlag(flags, "--policy-http-allowed-methods", policy.AllowedMethods)
	addManagedSliceFlag(flags, "--policy-http-allowed-hosts", policy.AllowedHosts)
	addManagedSliceFlag(flags, "--policy-http-blocked-hosts", policy.BlockedHosts)
	addManagedBoolFlag(flags, "--policy-http-allow-localhost", policy.AllowLocalhost)
	addManagedBoolFlag(flags, "--policy-http-allow-private-networks", policy.AllowPrivateNetworks)
	addManagedBoolFlag(flags, "--policy-http-allow-link-local", policy.AllowLinkLocal)

	if policy.DefaultHeaders != nil {
		flags["--policy-http-default-headers"] = struct{}{}
	}

	addManagedSliceFlag(flags, "--policy-http-blocked-request-headers", policy.BlockedRequestHeaders)

	if policy.Timeout != nil || policy.NoTimeout != nil {
		flags["--policy-http-timeout"] = struct{}{}
		flags["--policy-http-no-timeout"] = struct{}{}
	}

	if policy.MaxRequestSize != nil || policy.UnlimitedRequestSize != nil {
		flags["--policy-http-max-request-size"] = struct{}{}
		flags["--policy-http-unlimited-request-size"] = struct{}{}
	}

	if policy.MaxResponseSize != nil || policy.UnlimitedResponseSize != nil {
		flags["--policy-http-max-response-size"] = struct{}{}
		flags["--policy-http-unlimited-response-size"] = struct{}{}
	}

	if policy.MaxResponseHeaderSize != nil {
		flags["--policy-http-max-response-header-size"] = struct{}{}
	}

	if policy.FollowRedirects != nil {
		flags["--policy-http-follow-redirects"] = struct{}{}
	}

	if policy.MaxRedirects != nil {
		flags["--policy-http-max-redirects"] = struct{}{}
	}

	return flags
}
