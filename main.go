package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/MontFerret/lab/cmd"
	"github.com/urfave/cli/v2"
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

	if err := app.RunContext(ctx, os.Args); err != nil {
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

func newApp(version string, out io.Writer, errOut io.Writer) *cli.App {
	return &cli.App{
		Name:        "lab",
		Usage:       "run FQL test scripts",
		Description: "Ferret test runner",
		HideVersion: true,
		UsageText:   "lab [command] [command options]",
		Writer:      out,
		ErrWriter:   errOut,
		Action:      cmd.DefaultCommand,
		Flags:       cmd.RunFlags(true),
		Commands: []*cli.Command{
			cmd.RunCommand(),
			cmd.VersionCommand(version),
		},
	}
}
