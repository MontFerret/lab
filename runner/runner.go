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
		Runtime       runtime.Runtime
		PoolSize      uint64
		TestTimeout   time.Duration
		Times         uint64
		TimesInterval uint64
	}

	Runner struct {
		runtime      runtime.Runtime
		poolSize     uint64
		testTimeout  time.Duration
		testCount    uint64
		testInterval uint64
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
		runtime:      opts.Runtime,
		poolSize:     poolSize,
		testTimeout:  testTimeout,
		testCount:    times,
		testInterval: opts.TimesInterval,
	}, nil
}

func (r *Runner) Run(ctx Context, src sources.Source) Stream {
	onProgress := make(chan Result)
	onSummary := make(chan Summary)

	go func() {
		var failed int
		var passed int
		var sumDuration time.Duration

		onNext, onError := src.Read(ctx)

		for res := range r.consume(ctx, onNext, onError) {
			if res.Error != nil {
				failed++
			} else {
				passed++
			}

			sumDuration += res.Duration
			onProgress <- res
		}

		close(onProgress)

		onSummary <- Summary{
			Passed:   passed,
			Failed:   failed,
			Duration: sumDuration,
		}

		close(onSummary)
	}()

	return Stream{
		Progress: onProgress,
		Summary:  onSummary,
	}
}

func (r *Runner) consume(ctx Context, onNext <-chan sources.File, onError <-chan sources.Error) <-chan Result {
	out := make(chan Result)

	go func() {
		pool := NewPool(r.poolSize)
		var wg sync.WaitGroup
		var done bool

		for !done {
			select {
			case file, open := <-onNext:
				if !open {
					done = true

					break
				}

				f := file
				wg.Add(1)

				params := ctx.Params()
				params = params.Clone()

				pool.Go(func() {
					if ctx.Err() == nil {
						out <- r.runCase(ctx, f, params)
					}

					wg.Done()
				})
			case err, open := <-onError:
				if !open {
					done = true
					break
				}

				out <- Result{
					Times:    0,
					Filename: err.Filename,
					Duration: 0,
					Error:    errors.New(err.Message),
				}
			}
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
			Duration: 0,
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

		// we pause only if it's not the first execution
		if counter > 1 && r.testInterval > 0 {
			<-time.After(time.Duration(r.testInterval) * time.Second)
		}

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
