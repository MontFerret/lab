package testing

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type Unit struct {
	file    sources.File
	timeout time.Duration
}

func NewUnit(opts Options) (*Unit, error) {
	return &Unit{file: opts.File, timeout: opts.Timeout}, nil
}

func (unit *Unit) Run(ctx context.Context, rt runtime.Runtime, params Params) error {
	ctx, cancel := context.WithTimeout(ctx, unit.timeout)
	defer cancel()

	_, err := rt.Run(ctx, string(unit.file.Content), params.ToMap())

	if unit.mustFail() {
		if err != nil {
			return nil
		}

		return errors.New("expected to fail")
	}

	return err
}

func (unit *Unit) mustFail() bool {
	return strings.HasSuffix(unit.file.Name, ".fail.fql")
}
