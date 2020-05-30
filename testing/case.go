package testing

import (
	"context"
	"github.com/pkg/errors"
	"path"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type Case interface {
	Run(ctx context.Context, rt runtime.Runtime, params Params) error
}

func New(file sources.File) (Case, error) {
	if file.Error != nil {
		return nil, file.Error
	}

	switch path.Ext(file.Name) {
	case ".fql":
		return NewUnit(file)
	case ".yaml", ".yml":
		return NewSuite(file)
	default:
		return nil, errors.New("unknown file type")
	}
}
