package runtime

import (
	"context"
	"errors"
	"net/url"

	pkgerrors "github.com/pkg/errors"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

type (
	// FileSystemPolicy configures the sandboxed filesystem used by the built-in runtime.
	FileSystemPolicy struct {
		Root     string
		ReadOnly bool
	}

	Options struct {
		Type   string
		Params map[string]any
		// FSPolicy configures filesystem access for the built-in runtime only.
		FSPolicy *FileSystemPolicy
		// HTTPPolicy configures outbound HTTP for the built-in runtime only.
		HTTPPolicy []ferrethttp.PolicyOption
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
		return newBuiltin(params, opts.FSPolicy, opts.HTTPPolicy...)
	}

	u, err := url.Parse(opts.Type)

	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to parse remote runtime url")
	}

	switch u.Scheme {
	case "http", "https":
		if opts.FSPolicy != nil {
			return nil, errors.New("filesystem policy options are only supported by the built-in runtime")
		}

		if len(opts.HTTPPolicy) > 0 {
			return nil, errors.New("HTTP policy options are only supported by the built-in runtime")
		}

		return NewRemote(opts.Type, params)
	case "bin":
		if opts.FSPolicy != nil {
			return nil, errors.New("filesystem policy options are only supported by the built-in runtime")
		}

		if len(opts.HTTPPolicy) > 0 {
			return nil, errors.New("HTTP policy options are only supported by the built-in runtime")
		}

		return NewBinary(u.Host+u.Path, params)
	default:
		return newBuiltin(params, opts.FSPolicy, opts.HTTPPolicy...)
	}
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
