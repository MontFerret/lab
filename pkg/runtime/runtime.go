package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/MontFerret/api"
)

// Options configures runtime selection and adapter-specific execution values.
type Options struct {
	Type   string
	Params map[string]any
	// Endpoint and ConnectTimeout configure Wire transport establishment.
	Endpoint          string
	EndpointSet       bool
	ConnectTimeout    time.Duration
	ConnectTimeoutSet bool
	// FSPolicy configures filesystem access for the built-in runtime.
	FSPolicy *FileSystemPolicy
	// HTTPPolicy configures outbound HTTP for the built-in runtime.
	HTTPPolicy *HTTPPolicy
}

// New selects and configures a runtime. The context bounds Wire construction;
// callers supply independent execution contexts and close the runtime after runs settle.
func New(ctx context.Context, opts Options) (api.Runtime, error) {
	if isWireType(opts.Type) {
		if opts.ConnectTimeout == 0 && !opts.ConnectTimeoutSet {
			opts.ConnectTimeout = DefaultConnectTimeout
		}

		if err := validateWireOptions(opts); err != nil {
			return nil, err
		}

		rt, err := newWire(ctx, opts)
		if err != nil {
			return nil, err
		}

		return rt, nil
	}

	if opts.EndpointSet || opts.Endpoint != "" || opts.ConnectTimeoutSet || opts.ConnectTimeout != 0 {
		return nil, errors.New("--runtime-endpoint and --runtime-connect-timeout require --runtime wire")
	}

	params := opts.Params

	if params == nil {
		params = make(map[string]any)
	}

	if opts.Type == "" {
		return newConfiguredBuiltin(params, opts.FSPolicy, opts.HTTPPolicy)
	}

	u, err := url.Parse(opts.Type)

	if err != nil {
		return nil, fmt.Errorf("failed to parse remote runtime url: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		if opts.FSPolicy.hasSettings() {
			return nil, errors.New("filesystem policy options are not supported by HTTP runtimes")
		}

		if opts.HTTPPolicy.hasSettings() {
			return nil, errors.New("HTTP policy options are not supported by HTTP runtimes")
		}

		rt, err := NewRemote(opts.Type, params)
		if err != nil {
			return nil, err
		}

		return rt, nil
	case "bin":
		return nil, errors.New("binary runtimes are no longer supported; use the built-in, HTTP, or Wire runtime")
	default:
		return newConfiguredBuiltin(params, opts.FSPolicy, opts.HTTPPolicy)
	}
}

func newConfiguredBuiltin(params map[string]any, fsPolicy *FileSystemPolicy, httpPolicy *HTTPPolicy) (api.Runtime, error) {
	if err := fsPolicy.validate(); err != nil {
		return nil, err
	}

	options, err := httpPolicy.validatedFerretOptions()
	if err != nil {
		return nil, fmt.Errorf("HTTP policy: %w", err)
	}

	rt, err := newBuiltin(params, fsPolicy, options...)
	if err != nil {
		return nil, err
	}

	return rt, nil
}
