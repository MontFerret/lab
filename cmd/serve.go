package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"
)

func ServeCommand() *cli.Command {
	return &cli.Command{
		Name:      "serve",
		Usage:     "Serve one or more local HTTP services",
		UsageText: "lab serve [options]",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "static",
				Usage:   "served directory mapping (<path>, <path>:<port>, <path>@<alias>, <path>@<alias>:<port>)",
				Sources: cli.EnvVars("LAB_STATIC"),
			},
			&cli.StringSliceFlag{
				Name:    "mock",
				Usage:   "OpenAPI mock API spec mapping (<path>, <path>:<port>, <path>@<alias>, <path>@<alias>:<port>)",
				Sources: cli.EnvVars("LAB_MOCK"),
			},
			&cli.StringFlag{
				Name:    "serve-bind",
				Usage:   "host to bind local servers to (host only, no port)",
				Sources: cli.EnvVars("LAB_SERVE_BIND"),
			},
			&cli.StringFlag{
				Name:    "serve-host",
				Usage:   "host to advertise for local server URLs (host only, no port)",
				Sources: cli.EnvVars("LAB_SERVE_HOST"),
			},
		},
		Action: ServeAction,
	}
}

func ServeAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() > 0 {
		return cli.Exit("serve entries must use --static or --mock", 1)
	}

	staticValues := cmd.StringSlice("static")
	mockAPIValues := cmd.StringSlice("mock")

	if len(staticValues) == 0 && len(mockAPIValues) == 0 {
		if err := showCurrentCommandHelp(cmd); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	staticEntries, err := toServeEntries(staticValues)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	mockAPIEntries, err := toMockAPIEntries(mockAPIValues)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	staticManager, err := createStaticServerManagerFromCommand(cmd, staticEntries)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	mockManager, err := createMockAPIServerManagerFromCommand(cmd, mockAPIEntries)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	if staticManager != nil {
		if err := staticManager.Start(ctx); err != nil {
			return cli.Exit(err.Error(), 1)
		}
	}

	if mockManager != nil {
		if err := mockManager.Start(ctx); err != nil {
			if staticManager != nil {
				stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				_ = staticManager.Stop(stopCtx)
				cancel()
			}

			return cli.Exit(err.Error(), 1)
		}
	}

	w := appWriter(cmd)

	if staticManager != nil {
		for name, addr := range staticManager.Endpoints() {
			fmt.Fprintf(w, "Serving %q at %s\n", name, addr)
		}
	}

	if mockManager != nil {
		for name, addr := range mockManager.Endpoints() {
			fmt.Fprintf(w, "Serving mock API %q at %s\n", name, addr)
		}
	}

	// Block until context is cancelled (e.g. Ctrl+C)
	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var stopErrs []error
	if mockManager != nil {
		stopErrs = append(stopErrs, mockManager.Stop(stopCtx))
	}

	if staticManager != nil {
		stopErrs = append(stopErrs, staticManager.Stop(stopCtx))
	}

	return errors.Join(stopErrs...)
}
