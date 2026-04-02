package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
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
				Sources: cli.EnvVars("LAB_RUNTIME"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			rt, err := newRuntime(cmd, nil)

			if err != nil {
				return err
			}

			rtVersion, err := rt.Version(ctx)

			if err != nil {
				return err
			}

			fmt.Fprintln(appWriter(cmd), "Version")
			fmt.Fprintf(appWriter(cmd), "  Self: %s\n", self)
			fmt.Fprintf(appWriter(cmd), "  Runtime: %s\n", rtVersion)

			return nil
		},
	}
}
