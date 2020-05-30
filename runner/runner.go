package runner

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
	"github.com/MontFerret/lab/testing"
)

type (
	Options struct {
		Runtime     runtime.Runtime
		PoolSize    uint64
		TestTimeout time.Duration
	}

	Runner struct {
		runtime     runtime.Runtime
		poolSize    uint64
		testTimeout time.Duration
	}
)

func New(opts Options) (*Runner, error) {
	if opts.Runtime == nil {
		return nil, errors.New("missed runtime")
	}

	poolSize := opts.PoolSize

	if poolSize == 0 {
		poolSize = 1
	}

	testTimeout := opts.TestTimeout

	if testTimeout == 0 {
		testTimeout = time.Second * 30
	}

	return &Runner{
		runtime:     opts.Runtime,
		poolSize:    poolSize,
		testTimeout: testTimeout,
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

		for res := range r.runTests(ctx, stream.Files) {
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

func (r *Runner) runTests(ctx Context, files <-chan sources.File) <-chan Result {
	out := make(chan Result)

	go func() {
		pool := NewPool(r.poolSize)
		var wg sync.WaitGroup

		for f := range files {
			f := f
			wg.Add(1)

			params := ctx.Params()
			params = params.Clone()

			pool.Go(func() {
				if ctx.Err() == nil {
					out <- r.runCase(ctx, f, params)
				}

				wg.Done()
			})
		}

		wg.Wait()

		close(out)
	}()

	return out
}

func (r *Runner) runCase(ctx context.Context, file sources.File, params testing.Params) Result {
	testCase, err := testing.New(file)

	if err != nil {
		return Result{
			Filename: file.Name,
			Duration: time.Duration(0) * time.Millisecond,
			Error:    err,
		}
	}

	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, r.testTimeout)
	defer cancel()

	err = testCase.Run(ctx, r.runtime, params)

	duration := time.Since(start)

	return Result{
		Filename: file.Name,
		Duration: duration,
		Error:    err,
	}
}
