package runner

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
	"github.com/pkg/errors"
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
		case e, ok := <-stream.Error:
			if e != nil {
				err = e
				break
			}

			if !ok {
				done = ok
				break
			}
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
	if file.Error != nil {
		return Result{
			Filename: file.Name,
			Duration: time.Duration(0) * time.Millisecond,
			Error:    file.Error,
		}
	}

	mustFail := mustFail(file.Name)
	start := time.Now()

	out, err := r.runtime.Run(ctx, string(file.Content), ctx.params)

	duration := time.Since(start)

	if err != nil {
		if mustFail {
			return Result{
				Filename: file.Name,
				Duration: duration,
			}
		}

		return Result{
			Filename: file.Name,
			Duration: duration,
			Error:    errors.Wrap(err, "failed to execute query"),
		}
	}

	if mustFail {
		return Result{
			Filename: file.Name,
			Duration: duration,
			Error:    errors.New("expected to fail"),
		}
	}

	var result string

	if err := json.Unmarshal(out, &result); err != nil {
		return Result{
			Filename: file.Name,
			Duration: duration,
			Error:    err,
		}
	}

	if result == "" {
		return Result{
			Filename: file.Name,
			Duration: duration,
		}
	}

	return Result{
		Filename: file.Name,
		Duration: duration,
		Error:    errors.New(result),
	}
}
