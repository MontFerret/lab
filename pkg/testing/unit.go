package testing

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
)

const legacyExpectedFailureWarning = "'.fail.fql' expected-failure tests are deprecated; use a YAML test with 'expect.error' instead"

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

	_, err := rt.Run(ctx, ferret.NewSource(unit.file.Name, string(unit.file.Content)), params.ToMap())

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

// DeprecationWarning reports the compatibility warning for legacy expected-failure units.
func (unit *Unit) DeprecationWarning() string {
	if unit.mustFail() {
		return legacyExpectedFailureWarning
	}

	return ""
}
