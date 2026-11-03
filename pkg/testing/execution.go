package testing

import (
	"context"
	"errors"

	"github.com/MontFerret/api"
)

// Side-effect-only tests must observe terminal errors even when they discard every byte.
func runForEffect(ctx context.Context, rt api.Runtime, source api.Source, params map[string]any) error {
	out, err := rt.Run(ctx, source, api.WithParams(params))
	if err != nil {
		return err
	}

	if err := out.Consume(ctx, func(context.Context, []byte) error { return nil }); err != nil {
		return settleFailedOutput(out, err)
	}

	return nil
}

// Cancellation can reject consumption before the output takes ownership of cleanup.
func settleFailedOutput(out api.Output, err error) error {
	if closeErr := out.Close(); closeErr != nil && !errors.Is(err, closeErr) {
		return errors.Join(err, closeErr)
	}

	return err
}
