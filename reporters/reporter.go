package reporters

import (
	"context"
	"github.com/MontFerret/lab/runner"
)

type Reporter interface {
	Report(ctx context.Context, stream runner.Stream) error
}
