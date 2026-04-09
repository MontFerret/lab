package reporters

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/MontFerret/lab/v2/pkg/runner"
)

type Simple struct {
	out io.Writer
}

func NewSimple(out io.Writer) *Simple {
	return &Simple{out: out}
}

func (s *Simple) Report(ctx context.Context, stream runner.Stream) error {
	for res := range stream.Progress {
		if res.Error != nil {
			fmt.Fprintf(s.out, "FAIL file=%q duration=%s attempts=%d times=%d error=%q\n", res.Filename, res.Duration, res.Attempts, res.Times, res.Error.Error())
			continue
		}

		fmt.Fprintf(s.out, "PASS file=%q duration=%s attempts=%d times=%d\n", res.Filename, res.Duration, res.Attempts, res.Times)
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	case sum := <-stream.Summary:
		fmt.Fprintf(s.out, "DONE passed=%d failed=%d duration=%s\n", sum.Passed, sum.Failed, sum.Duration)

		if sum.HasErrors() {
			return errors.New("has errors")
		}

		return nil
	}
}
