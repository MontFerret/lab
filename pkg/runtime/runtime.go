package runtime

import (
	"context"
	"errors"
	"net/url"

	pkgerrors "github.com/pkg/errors"

	"github.com/MontFerret/ferret/v2/pkg/source"
)

type (
	// Options configures runtime selection and adapter-specific execution values.
	Options struct {
		Type   string
		Params map[string]any
		// FSPolicy configures filesystem access for built-in and binary runtimes.
		FSPolicy *FileSystemPolicy
		// HTTPPolicy configures outbound HTTP for built-in and binary runtimes.
		HTTPPolicy *HTTPPolicy
		// BinaryFlags contains additional arguments for the Ferret CLI run command.
		BinaryFlags []string
	}

	Func func(ctx context.Context, query *source.Source, params map[string]interface{}) ([]byte, error)

	Runtime interface {
		Version(ctx context.Context) (string, error)

		Run(ctx context.Context, query *source.Source, params map[string]interface{}) ([]byte, error)

		// Close releases resources owned by the runtime after all runs finish.
		Close() error
	}

	FuncStruct struct {
		fn Func
	}
)

func New(opts Options) (Runtime, error) {
	params := opts.Params

	if params == nil {
		params = make(map[string]any)
	}

	if opts.Type == "" {
		return newConfiguredBuiltin(params, opts.FSPolicy, opts.HTTPPolicy)
	}

	u, err := url.Parse(opts.Type)

	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to parse remote runtime url")
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
		return nil, pkgerrors.Wrap(err, "HTTP policy")
	}

	return newBuiltin(params, fsPolicy, options...)
}

func binaryPath(u *url.URL) string {
	if u.Opaque != "" {
		return u.Opaque
	}

	return u.Host + u.Path
}

func AsFunc(fn Func) Runtime {
	return &FuncStruct{fn}
}

func (f FuncStruct) Version(_ context.Context) (string, error) {
	return version, nil
}

func (f FuncStruct) Run(ctx context.Context, query *source.Source, params map[string]interface{}) ([]byte, error) {
	return f.fn(ctx, query, params)
}

func (f FuncStruct) Close() error {
	return nil
}
