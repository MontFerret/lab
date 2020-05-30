package sources

import (
	"context"
	"net/url"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
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

	var query = u.Query()
	filter := query.Get("filter")

	switch u.Scheme {
	case "file":
		fullPath := filepath.Join(u.Host, u.Path)
		parent := query.Get("from")

		if parent != "" {
			if !filepath.IsAbs(fullPath) {
				parentDir := filepath.Dir(parent)

				fp, err := filepath.Abs(filepath.Join(parentDir, u.Host, u.Path))

				if err != nil {
					return nil, err
				}

				fullPath = fp
			}
		}

		return NewFileSystem(fullPath, filter)
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
