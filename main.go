package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/MontFerret/lab/reporters"
	"github.com/MontFerret/lab/runner"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

func main() {
	app := &cli.App{
		Name:  "lab",
		Usage: "run FQL scripts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "cdp",
				Value:   "http://127.0.0.1:9222",
				Usage:   "Chrome DevTools Protocol address",
				EnvVars: []string{"FERRET_LAB_CDP"},
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "filter test files",
				Value:   "",
				EnvVars: []string{"FERRET_LAB_FILTER"},
			},
			&cli.StringFlag{
				Name:    "reporter",
				Aliases: []string{"r"},
				Usage:   "reporter (console, simple)",
				EnvVars: []string{"FERRET_LAB_REPORTER"},
				Value:   "console",
			},
			&cli.StringFlag{
				Name:    "runtime",
				Usage:   "url to remote Ferret runtime",
				EnvVars: []string{"FERRET_LAB_RUNTIME"},
			},
		},
		Action: func(c *cli.Context) error {
			r, err := runner.New(runtime.New(runtime.Options{
				RemoteURL: c.String("runtime"),
				CDP:       c.String("cdp"),
			}))

			if err != nil {
				return cli.Exit(err, 1)
			}

			var locations []string

			if c.NArg() == 0 {
				wd, err := os.Getwd()

				if err != nil {
					return cli.Exit(err, 1)
				}

				locations = []string{wd}
			} else {
				locations = c.Args().Slice()
			}

			src, err := sources.New(locations...)

			if err != nil {
				fmt.Println(err)

				os.Exit(1)
			}

			stream := r.Run(runner.NewContext(c.Context, nil), src)

			err = reporters.
				NewConsole(os.Stdout).
				Report(c.Context, stream)

			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		fmt.Println("failed to start the app")

		os.Exit(1)
	}
}
