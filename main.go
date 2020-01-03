package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/MontFerret/lab/reporters"
	"github.com/MontFerret/lab/runner"
)

func main() {
	app := &cli.App{
		Name:  "lab",
		Usage: "run FQL scripts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "cdp",
				Value: "http://127.0.0.1:9222",
				Usage: "Chrome DevTools Protocol address",
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "filter test files",
				Value:   "",
			},
		},
		Action: func(c *cli.Context) error {
			r, err := runner.New(runner.Settings{
				CDPAddress: c.String("cdp"),
				Filter:     "",
				Params:     nil,
			})

			if err != nil {
				return cli.Exit(err, 1)
			}

			var path []string

			if c.NArg() == 0 {
				wd, err := os.Getwd()

				if err != nil {
					return cli.Exit(err, 1)
				}

				path = []string{wd}
			} else {
				path = c.Args().Slice()
			}

			err = reporters.
				NewConsole(os.Stdout).
				Report(c.Context, r.Run(c.Context, path))

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
