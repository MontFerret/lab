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
		Times       uint64
	}

	Runner struct {
		runtime     runtime.Runtime
		poolSize    uint64
		testTimeout time.Duration
		testCount   uint64
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

	times := opts.Times

	if times == 0 {
		times = 1
	}

	testTimeout := opts.TestTimeout

	if testTimeout == 0 {
		testTimeout = time.Second * 30
	}

	return &Runner{
		runtime:     opts.Runtime,
		poolSize:    poolSize,
		testTimeout: testTimeout,
		testCount:   times,
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

		for res := range r.runTests(ctx, stream.OnNext()) {
			if res.Error != nil {
				failed++
			} else {
				passed++
			}

			sumDuration += res.Duration
			onProgress <- res
		}

		close(onProgress)

		for e := range stream.OnError() {
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
	testCase, err := testing.New(testing.Options{
		File:    file,
		Timeout: r.testTimeout,
	})

	if err != nil {
		return Result{
			Times:    0,
			Filename: file.Name,
			Duration: time.Duration(0) * time.Millisecond,
			Error:    err,
		}
	}

	counter := uint64(0)
	totalDuration := int64(0)

	for {
		if counter == r.testCount {
			break
		}

		counter++

		currentStart := time.Now()

		err = testCase.Run(ctx, r.runtime, params)

		totalDuration += time.Since(currentStart).Nanoseconds()

		if err != nil {
			break
		}
	}

	return Result{
		Times:    counter,
		Filename: file.Name,
		Duration: time.Duration(totalDuration / int64(counter)), // average duration
		Error:    err,
	}
}
