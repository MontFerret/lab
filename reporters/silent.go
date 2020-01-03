package reporters

import (
	"context"
	"github.com/MontFerret/lab/runner"
	"github.com/pkg/errors"
)

type Silent struct{}

func NewSilent() *Silent {
	return &Silent{}
}

func (c *Silent) Report(ctx context.Context, stream runner.Stream) error {
	var done bool
	var err error

	for {
		if done {
			break
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case sum, ok := <-stream.Summary:
			if !ok {
				done = true
				break
			}

			if sum.Failed > 0 {
				err = errors.Errorf("Failed")
			}
		case e := <-stream.Error:
			done = true
			err = e
		}
	}

	return err
}
