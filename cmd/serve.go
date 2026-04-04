package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"
)

func ServeCommand() *cli.Command {
	return &cli.Command{
		Name:      "serve",
		Usage:     "Serve one or more local directories over HTTP",
		UsageText: "lab serve [entries...]",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "serve",
				Usage:   "served directory mapping (<path>, <path>:<port>, <path>@<alias>, <path>@<alias>:<port>)",
				Sources: cli.EnvVars("LAB_SERVE"),
			},
			&cli.StringFlag{
				Name:    "serve-bind",
				Usage:   "host to bind static servers to (host only, no port)",
				Sources: cli.EnvVars("LAB_SERVE_BIND"),
			},
			&cli.StringFlag{
				Name:    "serve-host",
				Usage:   "host to advertise for static server URLs (host only, no port)",
				Sources: cli.EnvVars("LAB_SERVE_HOST"),
			},
		},
		Action: ServeAction,
	}
}

func ServeAction(ctx context.Context, cmd *cli.Command) error {
	values := cmd.StringSlice("serve")
	if cmd.NArg() > 0 {
		values = append(values, cmd.Args().Slice()...)
	}

	if len(values) == 0 {
		if err := showCurrentCommandHelp(cmd); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	entries, err := toServeEntries(values)
	if err != nil {
		return cli.Exit(err, 1)
	}

	manager, err := createStaticServerManagerFromCommand(cmd, entries)
	if err != nil {
		return cli.Exit(err, 1)
	}

	if err := manager.Start(ctx); err != nil {
		return cli.Exit(err, 1)
	}

	endpoints := manager.Endpoints()
	w := appWriter(cmd)

	for name, addr := range endpoints {
		fmt.Fprintf(w, "Serving %q at %s\n", name, addr)
	}

	// Block until context is cancelled (e.g. Ctrl+C)
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return manager.Stop(ctx)
}
