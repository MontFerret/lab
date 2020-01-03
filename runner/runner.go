package runner

import (
	"context"
	"time"

	"github.com/MontFerret/lab/runner/suites"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type Runner struct {
	runtime runtime.Runtime
}

func New(rt runtime.Runtime) (*Runner, error) {
	return &Runner{
		rt,
	}, nil
}

func (r *Runner) Run(ctx Context, src sources.Source) Stream {
	onProgress := make(chan Result)
	onSummary := make(chan Summary)
	onError := make(chan error)

	go func() {
		if err := r.runScripts(ctx, src, onProgress, onSummary); err != nil {
			onError <- err
		}

		close(onProgress)
		close(onSummary)
		close(onError)
	}()

	return Stream{
		Progress: onProgress,
		Summary:  onSummary,
		Error:    onError,
	}
}

func (r *Runner) runScripts(
	ctx Context,
	src sources.Source,
	onProgress chan<- Result,
	onSummary chan<- Summary,
) error {
	var failed int
	var passed int
	var sumDuration time.Duration
	var done bool
	var err error

	stream := src.Read(ctx)

	for !done {
		select {
		case <-ctx.Done():
			return context.Canceled
		case file, ok := <-stream.Files:
			if !ok {
				done = true
				break
			}

			res := r.runScript(ctx, file)

			if res.Error != nil {
				failed++
			} else {
				passed++
			}

			sumDuration += res.Duration

			onProgress <- res
		case e := <-stream.Error:
			err = e
			done = true

			break
		}
	}

	onSummary <- Summary{
		Passed:   passed,
		Failed:   failed,
		Duration: sumDuration,
		Error:    err,
	}

	return nil
}

func (r *Runner) runScript(ctx Context, file sources.File) Result {
	suite, err := suites.New(file)

	if err != nil {
		return Result{
			Filename: file.Name,
			Duration: time.Duration(0) * time.Millisecond,
			Error:    err,
		}
	}

	start := time.Now()

	err = suite.Run(ctx, r.runtime, ctx.params)

	duration := time.Since(start)

	return Result{
		Filename: file.Name,
		Duration: duration,
		Error:    err,
	}
}
