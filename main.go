package main

import (
	"context"
	"fmt"
	"github.com/MontFerret/lab/cmd"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
)

var (
	version string
)

func main() {
	app := &cli.App{
		Name:        "lab",
		Usage:       "run FQL test scripts",
		Description: "Ferret test runner",
		HideVersion: true,
		UsageText:   "lab [global options] [files...]",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "files",
				Aliases: []string{"f"},
				EnvVars: []string{"LAB_FILES"},
				Usage:   "location of FQL script files to run",
			},
			&cli.Uint64Flag{
				Name:        "timeout",
				Aliases:     []string{"t"},
				Usage:       "test timeout in seconds",
				EnvVars:     []string{"LAB_TIMEOUT"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				Value:       30,
				DefaultText: "",
				Destination: nil,
				HasBeenSet:  false,
			},
			&cli.StringFlag{
				Name:    "cdp",
				Value:   "http://127.0.0.1:9222",
				Usage:   "Chrome DevTools Protocol address",
				EnvVars: []string{"LAB_CDP"},
			},
			&cli.StringFlag{
				Name:    "reporter",
				Usage:   "reporter (console, simple)",
				EnvVars: []string{"LAB_REPORTER"},
				Value:   "console",
			},
			&cli.StringFlag{
				Name:    "runtime",
				Aliases: []string{"r"},
				Usage:   "url to remote Ferret runtime (http, https or bin)",
				EnvVars: []string{"LAB_RUNTIME"},
			},
			&cli.StringSliceFlag{
				Name:    "runtime-param",
				Aliases: []string{"rp"},
				Usage:   "params for remote Ferret runtime (--runtime-param=headers:{\"KeyId\": \"abcd\"} --runtime-param=path:\"/ferret\" })",
				EnvVars: []string{"LAB_RUNTIME_PARAM"},
			},
			&cli.Uint64Flag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "number of multiple tests to run at a time",
				EnvVars: []string{"LAB_CONCURRENCY"},
				Value:   1,
			},
			&cli.Uint64Flag{
				Name:    "times",
				Usage:   "number of times to run each test",
				EnvVars: []string{"LAB_TIMES"},
				Value:   1,
			},
			&cli.Uint64Flag{
				Name:    "attempts",
				Aliases: []string{"a"},
				Usage:   "number of times to re-run failed tests",
				EnvVars: []string{"LAB_ATTEMPTS"},
				Value:   1,
			},
			&cli.Uint64Flag{
				Name:    "times-interval",
				Usage:   "interval between test cycles in seconds",
				EnvVars: []string{"LAB_TIMES_INTERVAL"},
				Value:   0,
			},
			&cli.StringSliceFlag{
				Name:        "cdn",
				Usage:       "file or directory to serve via HTTP (./dir as default or ./dir@name with alias)",
				EnvVars:     []string{"LAB_CDN"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				TakesFile:   false,
				Value:       cli.NewStringSlice(),
				DefaultText: "",
				HasBeenSet:  false,
			},
			&cli.StringSliceFlag{
				Name:        "param",
				Aliases:     []string{"p"},
				Usage:       "query parameter (--param=foo:\"bar\", --param=id:1)",
				EnvVars:     []string{"LAB_PARAM"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				TakesFile:   false,
				Value:       nil,
				DefaultText: "",
				HasBeenSet:  false,
			},
			&cli.StringSliceFlag{
				Name:        "wait",
				Aliases:     []string{"w"},
				Usage:       "tests and waits on the availability of remote resources (--wait http://127.0.0.1:9222/json/version --wait postgres://locahost:5432/mydb)",
				EnvVars:     []string{"LAB_WAIT"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				TakesFile:   false,
				Value:       nil,
				DefaultText: "",
				HasBeenSet:  false,
			},
			&cli.Uint64Flag{
				Name:        "wait-timeout",
				Aliases:     []string{"wt"},
				Usage:       "wait timeout in seconds",
				EnvVars:     []string{"LAB_WAIT_TIMEOUT"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				Value:       5,
				DefaultText: "",
				Destination: nil,
				HasBeenSet:  false,
			},
			&cli.Uint64Flag{
				Name:        "wait-attempts",
				Aliases:     nil,
				Usage:       "wait attempts",
				EnvVars:     []string{"LAB_WAIT_ATTEMPTS"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				Value:       5,
				DefaultText: "",
				Destination: nil,
				HasBeenSet:  false,
			},
		},
		Action: cmd.DefaultCommand,
		Commands: []*cli.Command{
			cmd.VersionCommand(version),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, os.Kill)

	go func() {
		for {
			<-ch
			cancel()
		}
	}()

	defer cancel()

	if err := app.RunContext(ctx, os.Args); err != nil {
		if err == context.Canceled {
			fmt.Println("Terminated")
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}
