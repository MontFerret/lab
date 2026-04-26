package runtime

import (
	"context"
	"os"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

var version = "unknown"

type Builtin struct {
	engine *ferret.Engine
}

func NewBuiltin(params map[string]any) (*Builtin, error) {
	dir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	c, err := ferret.New(
		ferret.WithFSRoot(dir),
		ferret.WithParams(params),
	)

	if err != nil {
		return nil, err
	}

	return &Builtin{c}, nil
}

func (r *Builtin) Version(_ context.Context) (string, error) {
	return version, nil
}

func (r *Builtin) Run(ctx context.Context, query *source.Source, params map[string]any) ([]byte, error) {
	out, err := r.engine.Run(ctx, query, ferret.WithSessionParams(params))

	if err != nil {
		return nil, err
	}

	return out.Content, nil
}
