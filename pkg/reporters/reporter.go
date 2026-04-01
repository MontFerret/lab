package reporters

import (
	"context"

	"github.com/MontFerret/lab/v2/pkg/runner"
)

type Reporter interface {
	Report(ctx context.Context, stream runner.Stream) error
}
