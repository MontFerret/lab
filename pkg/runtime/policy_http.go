package runtime

import (
	"fmt"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

// HTTPPolicy describes outbound HTTP policy overrides for the built-in runtime.
// Nil fields leave the runtime's own setting unchanged.
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
