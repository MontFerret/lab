package runner

import (
	"sync"
	"time"

	"github.com/MontFerret/lab/runner/suites"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type Runner struct {
	runtime  runtime.Runtime
	poolSize uint64
}

func New(rt runtime.Runtime, poolSize uint64) (*Runner, error) {
	if poolSize == 0 {
		poolSize = 1
	}

	return &Runner{
		rt,
		poolSize,
	}, nil
}

func (r *Runner) Run(ctx Context, src sources.Source) Stream {
	onProgress := make(chan Result)
	onSummary := make(chan Summary)

	go func() {
		var failed int
		var passed int
		var sumDuration time.Duration
		errs := make([]error, 0, 5)

		stream := src.Read(ctx)

		for res := range r.runSuites(ctx, stream.Files) {
			if res.Error != nil {
				failed++
			} else {
				passed++
			}

			sumDuration += res.Duration
			onProgress <- res
		}

		close(onProgress)

		for e := range stream.Errors {
			errs = append(errs, e)
		}

		onSummary <- Summary{
			Passed:   passed,
			Failed:   failed,
			Duration: sumDuration,
			Errors:   errs,
		}

		close(onSummary)
	}()

	return Stream{
		Progress: onProgress,
		Summary:  onSummary,
	}
}

func (r *Runner) runSuites(ctx Context, files <-chan sources.File) <-chan Result {
	out := make(chan Result)

	go func() {
		pool := NewPool(r.poolSize)
		var wg sync.WaitGroup

		for f := range files {
			wg.Add(1)

			pool.Go(func() {
				if ctx.Err() == nil {
					out <- r.runSuite(ctx, f)
				}

				wg.Done()
			})
		}

		wg.Wait()

		close(out)
	}()

	return out
}

func (r *Runner) runSuite(ctx Context, file sources.File) Result {
	suite, err := suites.New(file)

	if err != nil {
		return Result{
			Filename: file.Name,
			Duration: time.Duration(0) * time.Millisecond,
			Error:    err,
		}
	}

	start := time.Now()

	err = suite.Run(ctx, r.runtime, ctx.Params().ToMap())

	duration := time.Since(start)

	return Result{
		Filename: file.Name,
		Duration: duration,
		Error:    err,
	}
}
