package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-waitfor/waitfor"
	http "github.com/go-waitfor/waitfor-http"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/reporters"
	"github.com/MontFerret/lab/v2/pkg/runner"
	"github.com/MontFerret/lab/v2/pkg/sources"
	"github.com/MontFerret/lab/v2/pkg/testing"
)

func RunAction(ctx context.Context, cmd *cli.Command) error {
	locations, ok := locationsFromCommand(cmd)

	if !ok {
		if err := showCurrentCommandHelp(cmd); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	return runScripts(ctx, cmd, locations)
}

func RootAction(ctx context.Context, cmd *cli.Command) error {
	return cli.ShowAppHelp(cmd)
}

func RootUsageError(ctx context.Context, cmd *cli.Command, err error, _ bool) error {
	return showSubcommandUsageError(cmd, err)
}

func runScripts(ctx context.Context, cmd *cli.Command, locations []string) error {
	waitFor := cmd.StringSlice("wait")

	if len(waitFor) > 0 {
		wait := waitfor.New(
			http.Use(),
		)

		err := wait.Test(
			ctx,
			waitFor,
			waitfor.WithAttempts(cmd.Uint64("wait-attempts")),
			waitfor.WithInterval(cmd.Uint64("wait-timeout")),
		)

		if err != nil {
			return cli.Exit(errors.Wrap(err, "timeout"), 1)
		}
	}

	runtimeParams, err := toParams(cmd.StringSlice("runtime-param"))

	if err != nil {
		return cli.Exit(err, 1)
	}

	rt, err := newRuntime(cmd, runtimeParams)

	if err != nil {
		return cli.Exit(err, 1)
	}

	r, err := runner.New(runner.Options{
		Runtime:       rt,
		PoolSize:      cmd.Uint64("concurrency"),
		Attempts:      cmd.Uint64("attempts"),
		TestTimeout:   time.Duration(cmd.Uint64("timeout")) * time.Second,
		Times:         cmd.Uint64("times"),
		TimesInterval: cmd.Uint64("times-interval"),
	})

	if err != nil {
		return cli.Exit(err, 1)
	}

	src, err := sources.New(locations...)

	if err != nil {
		return cli.Exit(err, 1)
	}

	params := testing.NewParams()

	userParams, err := toParams(cmd.StringSlice("param"))

	if err != nil {
		return cli.Exit(err, 1)
	}

	params.SetUserValues(userParams)

	dirs, err := toDirectories(cmd.StringSlice("cdn"))

	if err != nil {
		return cli.Exit(err, 1)
	}

	cdnManager, err := createCDNManager(dirs)

	if err != nil {
		return cli.Exit(err, 1)
	}

	cdnNodes := cdnManager.Endpoints()
	cdnMap := make(map[string]string)
	params.SetSystemValue("cdn", cdnMap)

	for _, dir := range dirs {
		_, found := cdnMap[dir.Name]

		if found {
			return cli.Exit(errors.Errorf("directory name is already defined: %s", dir.Name), 1)
		}

		address, found := cdnNodes[dir.Name]

		if found {
			cdnMap[dir.Name] = address
		}
	}

	err = cdnManager.Start(ctx)

	if err != nil {
		return cli.Exit(errors.Wrap(err, "failed to start local server for CDN"), 1)
	}

	stream := r.Run(runner.NewContext(ctx, params), src)

	return reporters.NewConsole(appWriter(cmd)).
		Report(ctx, stream)
}

func showSubcommandUsageError(cmd *cli.Command, err error) error {
	fmt.Fprintf(appWriter(cmd), "Incorrect Usage: %s\n\n", err.Error())

	if helpErr := cli.ShowSubcommandHelp(cmd); helpErr != nil {
		return helpErr
	}

	return err
}

func showCurrentCommandHelp(cmd *cli.Command) error {
	templ := cmd.CustomHelpTemplate

	if templ == "" {
		templ = cli.CommandHelpTemplate
	}

	cli.HelpPrinter(appWriter(cmd), templ, cmd)

	return nil
}

func appWriter(cmd *cli.Command) io.Writer {
	if cmd != nil {
		root := cmd.Root()
		if root.Writer != nil {
			return root.Writer
		}
	}

	return os.Stdout
}
