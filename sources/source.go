package sources

import (
	"context"
	"github.com/pkg/errors"
	"net/url"
	"path"
	"path/filepath"
)

type Source interface {
	Read(ctx context.Context) Stream
}

func New(locations ...string) (Source, error) {
	switch len(locations) {
	case 0:
		return NewNoop(), nil
	case 1:
		return create(locations[0])
	default:
		a := NewAggregate()

		for _, loc := range locations {
			src, err := create(loc)

			if err != nil {
				return nil, err
			}

			a.Add(src)
		}

		return a, nil
	}
}

func create(location string) (Source, error) {
	u, err := url.Parse(location)

	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, errors.New("location scheme is not provided")
	}

	var filter string

	var query = u.Query()

	if query != nil {
		filter = query.Get("filter")
	}

	switch u.Scheme {
	case "file":
		return NewFileSystem(filepath.Join(u.Host, u.Path), filter)
	//case "http":
	//	return NewNoop(), nil
	case "git+http":
		return NewGit("http://"+path.Join(u.Host, u.Path), filter)
	case "git+https":
		return NewGit("https://"+path.Join(u.Host, u.Path), filter)
	default:
		return nil, errors.Errorf("unknown location provider: %s", u.Scheme)
	}
}
