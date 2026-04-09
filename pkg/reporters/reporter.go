package reporters

import (
	"context"
	"fmt"
	"io"

	"github.com/MontFerret/lab/v2/pkg/runner"
)

type Reporter interface {
	Report(ctx context.Context, stream runner.Stream) error
}

func New(name string, out io.Writer) (Reporter, error) {
	switch name {
	case "", "console":
		return NewConsole(out), nil
	case "simple":
		return NewSimple(out), nil
	default:
		return nil, fmt.Errorf("unknown reporter: %s", name)
	}
}
