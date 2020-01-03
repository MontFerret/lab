package sources

import (
	"context"
	"net/url"
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

	switch u.Scheme {
	case "file":
		path := filepath.Join(u.Host, u.Path)

		return NewFileSystem(path)
	case "http":
		return NewNoop(), nil
	case "git":
		return NewNoop(), nil
	default:
		return NewNoop(), nil
	}
}
