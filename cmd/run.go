package cmd

import (
	"context"
	"time"

	"github.com/go-waitfor/waitfor"
	http "github.com/go-waitfor/waitfor-http"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/reporters"
	"github.com/MontFerret/lab/v2/pkg/runner"
	"github.com/MontFerret/lab/v2/pkg/sources"
	"github.com/MontFerret/lab/v2/pkg/testing"
)

func RunCommand() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run FQL test scripts",
		UsageText: "lab run [options] [files...]",
		Flags:     RunFlags(false),
		Action:    RunAction,
	}
}

func RunFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "files",
			Aliases: []string{"f"},
			Sources: cli.EnvVars("LAB_FILES"),
			Usage:   "location of FQL script files to run",
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Usage:   "test timeout in seconds",
			Sources: cli.EnvVars("LAB_TIMEOUT"),
			Value:   30,
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "reporter",
			Usage:   "reporter (console, simple)",
			Sources: cli.EnvVars("LAB_REPORTER"),
			Value:   "console",
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "runtime",
			Aliases: []string{"r"},
			Usage:   "url to remote Ferret runtime (http, https or bin)",
			Sources: cli.EnvVars("LAB_RUNTIME"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "runtime-param",
			Aliases: []string{"rp"},
			Usage:   "params for remote Ferret runtime (--runtime-param=headers:{\"KeyId\": \"abcd\"} --runtime-param=path:\"/ferret\")",
			Sources: cli.EnvVars("LAB_RUNTIME_PARAM"),
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "concurrency",
			Aliases: []string{"c"},
			Usage:   "number of multiple tests to run at a time",
			Sources: cli.EnvVars("LAB_CONCURRENCY"),
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "times",
			Usage:   "number of times to run each test",
			Sources: cli.EnvVars("LAB_TIMES"),
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "attempts",
			Aliases: []string{"a"},
			Usage:   "number of times to re-run failed tests",
			Sources: cli.EnvVars("LAB_ATTEMPTS"),
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "times-interval",
			Usage:   "interval between test cycles in seconds",
			Sources: cli.EnvVars("LAB_TIMES_INTERVAL"),
			Value:   0,
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "serve",
			Usage:   "serve a local directory over HTTP during test execution (<path>, <path>:<port>, <path>@<alias>, <path>@<alias>:<port>)",
			Sources: cli.EnvVars("LAB_SERVE"),
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "serve-bind",
			Usage:   "host to bind static servers to (host only, no port)",
			Sources: cli.EnvVars("LAB_SERVE_BIND"),
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "serve-host",
			Usage:   "host to advertise for static server URLs (host only, no port)",
			Sources: cli.EnvVars("LAB_SERVE_HOST"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "param",
			Aliases: []string{"p"},
			Usage:   "query parameter (--param=foo:\"bar\", --param=id:1)",
			Sources: cli.EnvVars("LAB_PARAM"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "wait",
			Aliases: []string{"w"},
			Usage:   "tests and waits on the availability of remote resources (--wait http://127.0.0.1:9222/json/version --wait postgres://localhost:5432/mydb)",
			Sources: cli.EnvVars("LAB_WAIT"),
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "wait-timeout",
			Aliases: []string{"wt"},
			Usage:   "wait timeout in seconds",
			Sources: cli.EnvVars("LAB_WAIT_TIMEOUT"),
			Value:   5,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "wait-attempts",
			Usage:   "wait attempts",
			Sources: cli.EnvVars("LAB_WAIT_ATTEMPTS"),
			Value:   5,
			Hidden:  hidden,
		},
	}
}

func RunAction(ctx context.Context, cmd *cli.Command) error {
	locations, ok := locationsFromCommand(cmd)

	if !ok {
		if err := showCurrentCommandHelp(cmd); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	return runScripts(ctx, cmd, locations)
}

func runScripts(ctx context.Context, cmd *cli.Command, locations []string) error {
	waitFor := cmd.StringSlice("wait")

	if len(waitFor) > 0 {
		wait := waitfor.New(
			http.Use(),
		)

		err := wait.Test(
			ctx,
			waitFor,
			waitfor.WithAttempts(cmd.Uint64("wait-attempts")),
			waitfor.WithInterval(cmd.Uint64("wait-timeout")),
		)

		if err != nil {
			return cli.Exit(errors.Wrap(err, "timeout"), 1)
		}
	}

	runtimeParams, err := toParams(cmd.StringSlice("runtime-param"))

	if err != nil {
		return cli.Exit(err, 1)
	}

	rt, err := newRuntime(cmd, runtimeParams)

	if err != nil {
		return cli.Exit(err, 1)
	}

	r, err := runner.New(runner.Options{
		Runtime:       rt,
		PoolSize:      cmd.Uint64("concurrency"),
		Attempts:      cmd.Uint64("attempts"),
		TestTimeout:   time.Duration(cmd.Uint64("timeout")) * time.Second,
		Times:         cmd.Uint64("times"),
		TimesInterval: cmd.Uint64("times-interval"),
	})

	if err != nil {
		return cli.Exit(err, 1)
	}

	src, err := sources.New(locations...)

	if err != nil {
		return cli.Exit(err, 1)
	}

	params := testing.NewParams()

	userParams, err := toParams(cmd.StringSlice("param"))

	if err != nil {
		return cli.Exit(err, 1)
	}

	params.SetUserValues(userParams)

	serveEntries, err := toServeEntries(cmd.StringSlice("serve"))
	if err != nil {
		return cli.Exit(err, 1)
	}

	staticURLs := make(map[string]interface{})
	params.SetSystemValue("static", staticURLs)

	manager, err := createStaticServerManagerFromCommand(cmd, serveEntries)
	if err != nil {
		return cli.Exit(err, 1)
	}

	if manager != nil {
		if err := manager.Start(ctx); err != nil {
			return cli.Exit(err, 1)
		}

		defer func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = manager.Stop(stopCtx)
		}()

		for alias, address := range manager.Endpoints() {
			staticURLs[alias] = address
		}
	}

	stream := r.Run(runner.NewContext(ctx, params), src)

	reporter, err := reporters.New(cmd.String("reporter"), appWriter(cmd))
	if err != nil {
		return cli.Exit(err, 1)
	}

	return reporter.Report(ctx, stream)
}
