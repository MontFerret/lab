package runner

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"github.com/MontFerret/ferret/pkg/runtime"
	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/MontFerret/lab/assertions"
)

type Runner struct {
	settings Settings
	compiler *compiler.Compiler
}

func New(settings Settings) (*Runner, error) {
	c := compiler.New()

	ns := c.Namespace("T")

	if err := assertions.Assertions(ns); err != nil {
		return nil, err
	}

	return &Runner{
		settings,
		c,
	}, nil
}

func (r *Runner) Run(ctx context.Context, dirOrFile []string) Stream {
	ctx = drivers.WithContext(
		ctx,
		cdp.NewDriver(cdp.WithAddress(r.settings.CDPAddress)),
	)

	ctx = drivers.WithContext(
		ctx,
		http.NewDriver(),
		drivers.AsDefault(),
	)

	onProgress := make(chan Result)
	onSummary := make(chan Summary)
	onError := make(chan error)

	go func() {
		if err := r.runScripts(ctx, dirOrFile, onProgress, onSummary); err != nil {
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

func (r *Runner) runScripts(ctx context.Context, dir []string, onProgress chan<- Result, onSummary chan<- Summary) error {
	var filter glob.Glob

	if r.settings.Filter != "" {
		f, err := glob.Compile(r.settings.Filter)

		if err != nil {
			return err
		}

		filter = f
	}

	var failed int
	var passed int
	var sumDuration time.Duration

	for _, d := range dir {
		err := r.traverseDir(ctx, d, filter, func(name string) error {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
				var res Result

				b, err := ioutil.ReadFile(name)

				if err != nil {
					onProgress <- Result{
						File:  name,
						Error: errors.Wrap(err, "failed to read file"),
					}
				} else {
					res = r.runScript(ctx, name, string(b))
				}

				if res.Error != nil {
					failed++
				} else {
					passed++
				}

				sumDuration += res.Duration

				onProgress <- res
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	onSummary <- Summary{
		Passed:   passed,
		Failed:   failed,
		Duration: sumDuration,
	}

	return nil
}

func (r *Runner) runScript(ctx context.Context, name, script string) Result {
	start := time.Now()

	p, err := r.compiler.Compile(script)

	if err != nil {
		return Result{
			File:     name,
			Duration: time.Duration(0) * time.Millisecond,
			Error:    errors.Wrap(err, "failed to compile query"),
		}
	}

	mustFail := mustFail(name)

	out, err := p.Run(
		ctx,
		runtime.WithLog(zerolog.ConsoleWriter{Out: os.Stdout}),
		runtime.WithParams(r.settings.Params),
	)

	duration := time.Since(start)

	if err != nil {
		if mustFail {
			return Result{
				File:     name,
				Duration: duration,
			}
		}

		return Result{
			File:     name,
			Duration: duration,
			Error:    errors.Wrap(err, "failed to execute query"),
		}
	}

	if mustFail {
		return Result{
			File:     name,
			Duration: duration,
			Error:    errors.New("expected to fail"),
		}
	}

	var result string

	if err := json.Unmarshal(out, &result); err != nil {
		return Result{
			File:     name,
			Duration: duration,
			Error:    err,
		}
	}

	if result == "" {
		return Result{
			File:     name,
			Duration: duration,
		}
	}

	return Result{
		File:     name,
		Duration: duration,
		Error:    errors.New(result),
	}
}

func (r *Runner) traverseDir(ctx context.Context, path string, filter glob.Glob, iteratee func(name string) error) error {
	fi, err := os.Stat(path)

	if err != nil {
		return err
	}

	// if only a single file was given
	if fi.Mode().IsRegular() {
		name := filepath.Join(path, fi.Name())

		// if not matched, skip the file
		if filter != nil && !filter.Match(name) {
			return nil
		}

		if !isFQLFile(path) {
			return errors.New("invalid file")
		}

		return iteratee(path)
	}

	files, err := ioutil.ReadDir(path)

	if err != nil {
		return err
	}

	for _, file := range files {
		name := filepath.Join(path, file.Name())

		// if not matched, skip the file
		if filter != nil && !filter.Match(name) {
			continue
		}

		if file.IsDir() {
			if err := r.traverseDir(ctx, name, filter, iteratee); err != nil {
				return err
			}

			continue
		}

		if !isFQLFile(file.Name()) {
			continue
		}

		if err := iteratee(name); err != nil {
			return err
		}
	}

	return nil
}
