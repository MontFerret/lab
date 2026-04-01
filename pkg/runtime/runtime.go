package runtime

import (
	"context"
	"net/url"

	"github.com/pkg/errors"

	"github.com/MontFerret/ferret/v2/pkg/source"
)

type (
	Options struct {
		Type       string
		CDPAddress string
		Params     map[string]any
	}

	Func func(ctx context.Context, query *source.Source, params map[string]interface{}) ([]byte, error)

	Runtime interface {
		Version(ctx context.Context) (string, error)

		Run(ctx context.Context, query *source.Source, params map[string]interface{}) ([]byte, error)
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
		return NewBuiltin(opts.CDPAddress, params)
	}

	u, err := url.Parse(opts.Type)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse remote runtime url")
	}

	switch u.Scheme {
	case "http", "https":
		return NewRemote(opts.Type, params)
	case "bin":
		return NewBinary(u.Host+u.Path, opts.CDPAddress, params)
	default:
		return NewBuiltin(opts.CDPAddress, params)
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
