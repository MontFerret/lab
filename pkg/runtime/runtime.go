package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/MontFerret/ferret/v2"
)

type (
	// Options configures runtime selection and adapter-specific execution values.
	Options struct {
		Type   string
		Params map[string]any
		// Endpoint and ConnectTimeout configure Wire transport establishment.
		Endpoint          string
		EndpointSet       bool
		ConnectTimeout    time.Duration
		ConnectTimeoutSet bool
		// FSPolicy configures filesystem access for built-in and binary runtimes.
		FSPolicy *FileSystemPolicy
		// HTTPPolicy configures outbound HTTP for built-in and binary runtimes.
		HTTPPolicy *HTTPPolicy
		// BinaryFlags contains additional arguments for the Ferret CLI run command.
		// A nil slice means no binary configuration was supplied.
		BinaryFlags []string
	}

	Runtime interface {
		Version(ctx context.Context) (string, error)

		Run(ctx context.Context, query ferret.Source, params map[string]any) ([]byte, error)

		// Close releases resources owned by the runtime after all runs finish.
		Close() error
	}
)

// New selects and configures a runtime. The context bounds Wire construction;
// callers supply independent execution contexts and close the runtime after runs settle.
func New(ctx context.Context, opts Options) (Runtime, error) {
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
		if len(opts.BinaryFlags) > 0 {
			return nil, errors.New("binary flags are only supported by binary runtimes")
		}

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

		if len(opts.BinaryFlags) > 0 {
			return nil, errors.New("binary flags are only supported by binary runtimes")
		}

		return NewRemote(opts.Type, params)
	case "bin":
		return NewBinary(BinaryOptions{
			Path:       binaryPath(u),
			Params:     params,
			Flags:      opts.BinaryFlags,
			FSPolicy:   opts.FSPolicy,
			HTTPPolicy: opts.HTTPPolicy,
		})
	default:
		if len(opts.BinaryFlags) > 0 {
			return nil, errors.New("binary flags are only supported by binary runtimes")
		}

		return newConfiguredBuiltin(params, opts.FSPolicy, opts.HTTPPolicy)
	}
}

func newConfiguredBuiltin(params map[string]any, fsPolicy *FileSystemPolicy, httpPolicy *HTTPPolicy) (*Builtin, error) {
	if err := fsPolicy.validate(); err != nil {
		return nil, err
	}

	options, err := httpPolicy.validatedFerretOptions()
	if err != nil {
		return nil, fmt.Errorf("HTTP policy: %w", err)
	}

	return newBuiltin(params, fsPolicy, options...)
}

func binaryPath(u *url.URL) string {
	if u.Opaque != "" {
		return u.Opaque
	}

	return u.Host + u.Path
}
