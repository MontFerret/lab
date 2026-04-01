package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func VersionCommand(self string) *cli.Command {
	return &cli.Command{
		Name:      "version",
		Usage:     "Show Lab version",
		UsageText: "lab version [options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "runtime",
				Aliases: []string{"r"},
				Usage:   "url to remote Ferret runtime (http, https or bin)",
				EnvVars: []string{"LAB_RUNTIME"},
			},
		},
		Action: func(c *cli.Context) error {
			rt, err := newRuntime(c, nil)

			if err != nil {
				return err
			}

			rtVersion, err := rt.Version(c.Context)

			if err != nil {
				return err
			}

			fmt.Fprintln(appWriter(c), "Version")
			fmt.Fprintf(appWriter(c), "  Self: %s\n", self)
			fmt.Fprintf(appWriter(c), "  Runtime: %s\n", rtVersion)

			return nil
		},
	}
}
