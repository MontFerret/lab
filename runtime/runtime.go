package runtime

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
)

type (
	Options struct {
		RemoteURL  string
		CDPAddress string
		Params     map[string]interface{}
	}

	Func func(ctx context.Context, query string, params map[string]interface{}) ([]byte, error)

	Runtime interface {
		Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error)
	}

	FuncStruct struct {
		fn Func
	}
)

func New(opts Options) (Runtime, error) {
	if opts.RemoteURL == "" {
		return NewBuiltin(opts.CDPAddress, opts.Params)
	}

	u, err := url.Parse(opts.RemoteURL)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse remote runtime url")
	}

	switch u.Scheme {
	case "http", "https":
		return NewHTTP(opts.RemoteURL, opts.Params)
	case "bin":
		return NewBinary(u.Host+u.Path, opts.CDPAddress, opts.Params)
	default:
		return NewBinary(u.Host+u.Path, opts.CDPAddress, opts.Params)
	}
}

func AsFunc(fn Func) Runtime {
	return &FuncStruct{fn}
}

func (f FuncStruct) Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
	return f.fn(ctx, query, params)
}
