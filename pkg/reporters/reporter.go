package reporters

import (
	"context"

	"github.com/MontFerret/lab/pkg/runner"
)

type Reporter interface {
	Report(ctx context.Context, stream runner.Stream) error
}
