package sources

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
)

type (
	Source interface {
		Read(ctx context.Context) (onNext <-chan File, onError <-chan Error)
		Resolve(ctx context.Context, path string) (onNext <-chan File, onError <-chan Error)
	}

	SourceFactory func(u url.URL) (Source, error)

	SourceType int
)

const (
	SourceTypeUnknown SourceType = 0
	SourceTypeFS      SourceType = 1
	SourceTypeHTTP    SourceType = 2
	SourceTypeGIT     SourceType = 3
)

var typeByScheme = map[string]SourceType{
	"file":      SourceTypeFS,
	"http":      SourceTypeHTTP,
	"https":     SourceTypeHTTP,
	"git+http":  SourceTypeGIT,
	"git+https": SourceTypeGIT,
}

var factoryByType = map[SourceType]SourceFactory{
	SourceTypeFS:   NewFileSystem,
	SourceTypeHTTP: NewHTTP,
	SourceTypeGIT:  NewGit,
}

func New(locations ...string) (Source, error) {
	switch len(locations) {
	case 0:
		return NewNoop(), nil
	case 1:
		return Create(locations[0])
	default:
		a := NewAggregate()

		for _, loc := range locations {
			src, err := Create(loc)

			if err != nil {
				return nil, err
			}

			a.Add(src)
		}

		return a, nil
	}
}

func Create(str string) (Source, error) {
	u, err := url.Parse(str)

	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, errors.New("source scheme is not provided")
	}

	srcType := GetType(u)

	factory, found := factoryByType[srcType]

	if !found {
		return nil, errors.Errorf("unknown source provider: %s", u.Scheme)
	}

	return factory(*u)
}

func Resolve(file File, ref string) (Source, error) {
	u, err := url.Parse(ref)

	if err != nil {
		return nil, err
	}

	// set a query param that indicates from what relative location to resolve a given script
	q := u.Query()
	q.Add("from", file.Name)
	u.RawQuery = q.Encode()

	return Create(u.String())
}

func GetType(u *url.URL) SourceType {
	if u == nil {
		return SourceTypeUnknown
	}

	srcType, exists := typeByScheme[u.Scheme]

	if !exists {
		return SourceTypeUnknown
	}

	return srcType
}
