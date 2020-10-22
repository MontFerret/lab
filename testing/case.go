package testing

import (
	"context"
	"github.com/pkg/errors"
	"path"
	"time"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type (
	Options struct {
		File    sources.File
		Timeout time.Duration
	}

	Case interface {
		Run(ctx context.Context, rt runtime.Runtime, params Params) error
	}
)

func New(opts Options) (Case, error) {
	switch path.Ext(opts.File.Name) {
	case ".fql":
		return NewUnit(opts)
	case ".yaml", ".yml":
		return NewSuite(opts)
	default:
		return nil, errors.New("unknown file type")
	}
}
