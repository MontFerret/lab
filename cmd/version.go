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
				Usage:   "Ferret runtime (HTTP URL or bin:<Ferret CLI v2 path>)",
				Sources: cli.EnvVars("LAB_RUNTIME"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			rt, err := newRuntime(cmd, nil)

			if err != nil {
				return err
			}

			rtVersion, versionErr := rt.Version(ctx)
			closeErr := rt.Close()

			if versionErr != nil {
				return versionErr
			}

			if closeErr != nil {
				return closeErr
			}

			fmt.Fprintln(appWriter(cmd), "Version")
			fmt.Fprintf(appWriter(cmd), "  Self: %s\n", self)
			fmt.Fprintf(appWriter(cmd), "  Runtime: %s\n", rtVersion)

			return nil
		},
	}
}
