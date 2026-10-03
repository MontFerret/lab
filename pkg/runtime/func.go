package runtime

import (
	"context"

	"github.com/MontFerret/ferret/v2"
)

type (
	Func func(ctx context.Context, query ferret.Source, params map[string]any) ([]byte, error)

	FuncStruct struct {
		fn Func
	}
)

// AsFunc adapts an execution callback without owning resources.
func AsFunc(fn Func) Runtime {
	return &FuncStruct{fn}
}

func (f FuncStruct) Version(_ context.Context) (string, error) {
	return version, nil
}

func (f FuncStruct) Run(ctx context.Context, query ferret.Source, params map[string]any) ([]byte, error) {
	return f.fn(ctx, query, params)
}

func (f FuncStruct) Close() error {
	return nil
}
