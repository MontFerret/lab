package suites

import (
	"context"
	"errors"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
	"strings"
)

type FQL struct {
	file sources.File
}

func NewFQL(file sources.File) *FQL {
	return &FQL{file: file}
}

func (suite *FQL) Run(ctx context.Context, rt runtime.Runtime, params map[string]interface{}) error {
	_, err := rt.Run(ctx, string(suite.file.Content), params)

	if suite.mustFail() {
		if err != nil {
			return nil
		}

		return errors.New("expected to fail")
	}

	return err
}

func (suite *FQL) mustFail() bool {
	return strings.HasSuffix(suite.file.Name, ".fail.fql")
}
