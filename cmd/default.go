package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-waitfor/waitfor"
	http "github.com/go-waitfor/waitfor-http"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/MontFerret/lab/v2/pkg/reporters"
	runner2 "github.com/MontFerret/lab/v2/pkg/runner"
	"github.com/MontFerret/lab/v2/pkg/sources"
	"github.com/MontFerret/lab/v2/pkg/testing"
)

func RunAction(c *cli.Context) error {
	locations, ok := locationsFromContext(c)

	if !ok {
		if err := showCurrentCommandHelp(c); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	return runScripts(c, locations)
}

func RootAction(c *cli.Context) error {
	return cli.ShowAppHelp(c)
}

func RootUsageError(c *cli.Context, err error, _ bool) error {
	return showSubcommandUsageError(c, err)
}

func runScripts(c *cli.Context, locations []string) error {
	waitFor := c.StringSlice("wait")

	if len(waitFor) > 0 {
		wait := waitfor.New(
			http.Use(),
		)

		err := wait.Test(
			c.Context,
			waitFor,
			waitfor.WithAttempts(c.Uint64("wait-attempts")),
			waitfor.WithInterval(c.Uint64("wait-timeout")),
		)

		if err != nil {
			return cli.Exit(errors.Wrap(err, "timeout"), 1)
		}
	}

	runtimeParams, err := toParams(c.StringSlice("runtime-param"))

	if err != nil {
		return cli.Exit(err, 1)
	}

	rt, err := newRuntime(c, runtimeParams)

	if err != nil {
		return cli.Exit(err, 1)
	}

	r, err := runner2.New(runner2.Options{
		Runtime:       rt,
		PoolSize:      c.Uint64("concurrency"),
		Attempts:      c.Uint64("attempts"),
		TestTimeout:   time.Duration(c.Uint64("timeout")) * time.Second,
		Times:         c.Uint64("times"),
		TimesInterval: c.Uint64("times-interval"),
	})

	if err != nil {
		return cli.Exit(err, 1)
	}

	src, err := sources.New(locations...)

	if err != nil {
		return cli.Exit(err, 1)
	}

	params := testing.NewParams()

	userParams, err := toParams(c.StringSlice("param"))

	if err != nil {
		return cli.Exit(err, 1)
	}

	params.SetUserValues(userParams)

	dirs, err := toDirectories(c.StringSlice("cdn"))

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

	err = cdnManager.Start(c.Context)

	if err != nil {
		return cli.Exit(errors.Wrap(err, "failed to start local server for CDN"), 1)
	}

	stream := r.Run(runner2.NewContext(c.Context, params), src)

	return reporters.NewConsole(appWriter(c)).
		Report(c.Context, stream)
}

func showSubcommandUsageError(c *cli.Context, err error) error {
	fmt.Fprintf(appWriter(c), "Incorrect Usage: %s\n\n", err.Error())

	if helpErr := cli.ShowSubcommandHelp(c); helpErr != nil {
		return helpErr
	}

	return err
}

func showCurrentCommandHelp(c *cli.Context) error {
	command := commandFromParentContext(c)

	if command == nil {
		return cli.ShowSubcommandHelp(c)
	}

	templ := command.CustomHelpTemplate

	if templ == "" {
		templ = cli.CommandHelpTemplate
	}

	cli.HelpPrinter(appWriter(c), templ, command)

	return nil
}

func commandFromParentContext(c *cli.Context) *cli.Command {
	if c == nil || c.Command == nil {
		return nil
	}

	lineage := c.Lineage()

	if len(lineage) < 2 || lineage[1] == nil || lineage[1].App == nil {
		return nil
	}

	parent := lineage[1]
	commands := parent.App.Commands

	if parent.Command != nil && parent.Command.Subcommands != nil {
		commands = parent.Command.Subcommands
	}

	for _, command := range commands {
		if command.HasName(c.Command.Name) {
			return command
		}
	}

	return nil
}

func appWriter(c *cli.Context) io.Writer {
	if c != nil && c.App != nil && c.App.Writer != nil {
		return c.App.Writer
	}

	return os.Stdout
}
