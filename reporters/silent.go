package reporters

import (
	"context"
	"github.com/MontFerret/lab/runner"
)

type Silent struct{}

func NewSilent() *Silent {
	return &Silent{}
}

func (c *Silent) Report(ctx context.Context, stream runner.Stream) error {
	for range stream.Progress {
	}

	select {
	case <-ctx.Done():
		break
	case <-stream.Summary:
		break
	}

	return nil
}
