package reporters

import (
	"context"
	"errors"
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
		return context.Canceled
	case sum := <-stream.Summary:
		if sum.HasErrors() {
			return errors.New("has errors")
		}
	}

	return nil
}
