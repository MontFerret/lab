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
	Options struct {
		Type   string
		Params map[string]any
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
		return NewBuiltin(params, opts.HTTPPolicy...)
	}

	u, err := url.Parse(opts.Type)

	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to parse remote runtime url")
	}

	switch u.Scheme {
	case "http", "https":
		if len(opts.HTTPPolicy) > 0 {
			return nil, errors.New("HTTP policy options are only supported by the built-in runtime")
		}

		return NewRemote(opts.Type, params)
	case "bin":
		if len(opts.HTTPPolicy) > 0 {
			return nil, errors.New("HTTP policy options are only supported by the built-in runtime")
		}

		return NewBinary(u.Host+u.Path, params)
	default:
		return NewBuiltin(params, opts.HTTPPolicy...)
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
