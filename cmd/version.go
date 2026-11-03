package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/runtime"
)

func VersionCommand(self string) *cli.Command {
	return &cli.Command{
		Name:      "version",
		Usage:     "Show Lab version",
		UsageText: "lab version [options]",
		Flags:     runtimeFlags(false),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			options, err := runtimeOptionsFromCommand(cmd)
			if err != nil {
				return err
			}

			rt, err := runtime.New(ctx, options)
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
