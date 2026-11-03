package runtime

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

type (
	// Options keeps adapter configuration separate from invocation parameters.
	Options struct {
		Target         string
		AdapterOptions map[string]any
		// FSPolicy configures filesystem access for the built-in runtime.
		FSPolicy *FileSystemPolicy
		// HTTPPolicy configures outbound HTTP for the built-in runtime.
		HTTPPolicy *HTTPPolicy
	}

	runtimeConfiguration struct {
		kind              string
		url               *url.URL
		http              httpOptions
		endpoint          string
		connectTimeout    time.Duration
		httpPolicyOptions []ferrethttp.PolicyOption
	}
)

// Validate checks the complete configuration without connecting or starting resources.
func (o Options) Validate() error {
	_, err := o.resolve()

	return err
}

func (o Options) resolve() (runtimeConfiguration, error) {
	config := runtimeConfiguration{}

	if o.Target != "" && o.Target != "builtin" {
		u, err := url.Parse(o.Target)
		if err != nil {
			return config, errors.New("invalid --runtime target: expected an HTTP/HTTPS URL or wire+tcp://127.0.0.1:<port>")
		}

		config.kind, config.url = u.Scheme, u
		if config.kind == "bin" {
			return config, errors.New("binary runtimes are no longer supported; use the built-in, HTTP, or Wire runtime")
		}

		if config.kind != "http" && config.kind != "https" && config.kind != "wire+tcp" {
			return config, errors.New("unsupported --runtime target: use builtin, an HTTP/HTTPS URL, or wire+tcp://127.0.0.1:<port>")
		}
	}

	if config.kind == "" {
		if err := validateAdapterOptionNames(o.AdapterOptions, "built-in"); err != nil {
			return config, err
		}

		if err := o.FSPolicy.validate(); err != nil {
			return config, err
		}

		options, err := o.HTTPPolicy.validatedFerretOptions()
		if err != nil {
			return config, fmt.Errorf("HTTP policy: %w", err)
		}

		config.httpPolicyOptions = options

		return config, nil
	}

	adapter := "HTTP"

	if config.kind == "wire+tcp" {
		adapter = "Wire"
	}

	if o.FSPolicy.hasSettings() {
		return config, fmt.Errorf("filesystem policy options are not supported by %s runtimes", adapter)
	}

	if o.HTTPPolicy.hasSettings() {
		return config, fmt.Errorf("HTTP policy options are not supported by %s runtimes", adapter)
	}

	if config.kind == "wire+tcp" {
		u := config.url
		if u.Opaque != "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || strings.Contains(o.Target, "#") {
			return config, errors.New("invalid Wire target: credentials, paths, queries, and fragments are not supported")
		}

		address, err := wireAddress("tcp://" + u.Host)
		if err != nil {
			return config, err
		}

		config.endpoint = "tcp://" + address
		config.connectTimeout, err = wireConnectTimeout(o.AdapterOptions)

		return config, err
	}

	if err := validateHTTPURL(config.url); err != nil {
		return config, err
	}

	options, err := newHTTPOptions(o.Target, o.AdapterOptions)
	config.http = options

	return config, err
}
