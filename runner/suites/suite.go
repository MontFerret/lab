package suites

import (
	"context"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
	"github.com/pkg/errors"
)

type Suite interface {
	Run(ctx context.Context, rt runtime.Runtime, params map[string]interface{}) error
}

func New(file sources.File) (Suite, error) {
	if file.Error != nil {
		return nil, file.Error
	}

	switch file.Name {
	case ".fql":
		return NewFQL(file), nil
	default:
		return nil, errors.New("unknown file type")
	}
}
