package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func ServeCommand() *cli.Command {
	return &cli.Command{
		Name:      "serve",
		Usage:     "Serve static files via HTTP",
		UsageText: "lab serve [options] [directories...]",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "cdn",
				Usage:   "directory to serve via HTTP (./dir as default or ./dir@name with alias)",
				Sources: cli.EnvVars("LAB_CDN"),
			},
		},
		Action: ServeAction,
	}
}

func ServeAction(ctx context.Context, cmd *cli.Command) error {
	// Collect directories from both positional args and --cdn flag
	values := cmd.StringSlice("cdn")
	if cmd.NArg() > 0 {
		values = append(values, cmd.Args().Slice()...)
	}

	if len(values) == 0 {
		if err := showCurrentCommandHelp(cmd); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	dirs, err := toDirectories(values)
	if err != nil {
		return cli.Exit(err, 1)
	}

	cdnManager, err := createCDNManager(dirs)
	if err != nil {
		return cli.Exit(err, 1)
	}

	endpoints := cdnManager.Endpoints()
	w := appWriter(cmd)

	for name, addr := range endpoints {
		fmt.Fprintf(w, "Serving %q at %s\n", name, addr)
	}

	err = cdnManager.Start(ctx)
	if err != nil {
		return cli.Exit(errors.Wrap(err, "failed to start static server"), 1)
	}

	// Block until context is cancelled (e.g. Ctrl+C)
	<-ctx.Done()

	return cdnManager.Stop(context.Background())
}
