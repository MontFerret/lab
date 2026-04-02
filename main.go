package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/cmd"
)

var (
	version string
)

func main() {
	app := newApp(version, os.Stdout, os.Stderr)

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			<-ch
			cancel()
		}
	}()

	defer cancel()

	if err := app.Run(ctx, os.Args); err != nil {
		if err == context.Canceled {
			fmt.Fprintln(app.ErrWriter, "Terminated")
		} else if err.Error() != "" {
			fmt.Fprintln(app.ErrWriter, err)
		}

		exitCode := 1

		if codeErr, ok := err.(cli.ExitCoder); ok {
			exitCode = codeErr.ExitCode()
		}

		os.Exit(exitCode)
	}
}

func newApp(version string, out io.Writer, errOut io.Writer) *cli.Command {
	return &cli.Command{
		Name:         "lab",
		Usage:        "run FQL test scripts",
		Description:  "Ferret test runner",
		UsageText:    "lab [command] [command options]",
		Writer:       out,
		ErrWriter:    errOut,
		Action:       cmd.RootAction,
		OnUsageError: cmd.RootUsageError,
		Commands: []*cli.Command{
			cmd.RunCommand(),
			cmd.ServeCommand(),
			cmd.VersionCommand(version),
		},
	}
}
