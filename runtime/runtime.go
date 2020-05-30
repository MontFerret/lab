package runtime

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

type (
	Options struct {
		RemoteURL string
		CDP       string
		Params    map[string]interface{}
	}

	Runtime interface {
		Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error)
	}
)

func New(opts Options) (Runtime, error) {
	if opts.RemoteURL == "" {
		return NewBuiltin(opts.CDP, opts.Params)
	}

	u, err := url.Parse(opts.RemoteURL)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse remote runtime url")
	}

	switch u.Scheme {
	case "http", "https":
		return NewHTTP(opts.RemoteURL, opts.Params)
	case "bin":
		return NewBinary(u.Host+u.Path, opts.Params)
	default:
		return nil, fmt.Errorf("invalid remote url: %s", opts.RemoteURL)
	}
}
