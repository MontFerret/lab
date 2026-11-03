package runtime

import (
	"context"
	"errors"

	"github.com/MontFerret/api"
)

func pointerTo[T any](value T) *T {
	return &value
}

func runAndCollect(ctx context.Context, rt api.Runtime, src api.Source, opts ...api.SessionOption) (*api.Content, error) {
	out, err := rt.Run(ctx, src, opts...)
	if err != nil {
		return nil, err
	}

	content, err := out.Collect(ctx)
	if err != nil {
		if closeErr := out.Close(); closeErr != nil && !errors.Is(err, closeErr) {
			err = errors.Join(err, closeErr)
		}
	}

	return content, err
}
