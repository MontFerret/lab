package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func RootAction(_ context.Context, cmd *cli.Command) error {
	return cli.ShowAppHelp(cmd)
}

func RootUsageError(_ context.Context, cmd *cli.Command, err error, _ bool) error {
	return showSubcommandUsageError(cmd, err)
}

func showSubcommandUsageError(cmd *cli.Command, err error) error {
	fmt.Fprintf(appWriter(cmd), "Incorrect Usage: %s\n\n", err.Error())

	if helpErr := cli.ShowSubcommandHelp(cmd); helpErr != nil {
		return helpErr
	}

	return err
}
