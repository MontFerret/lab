package testing

import (
	"context"
	"errors"
	"strings"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type Unit struct {
	file sources.File
}

func NewUnit(file sources.File) (*Unit, error) {
	return &Unit{file: file}, nil
}

func (suite *Unit) Run(ctx context.Context, rt runtime.Runtime, params Params) error {
	_, err := rt.Run(ctx, string(suite.file.Content), params.ToMap())

	if suite.mustFail() {
		if err != nil {
			return nil
		}

		return errors.New("expected to fail")
	}

	return err
}

func (suite *Unit) mustFail() bool {
	return strings.HasSuffix(suite.file.Name, ".fail.fql")
}
