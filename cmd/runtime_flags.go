package cmd

import (
	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/runtime"
)

func runtimeConnectionFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "runtime-endpoint",
			Usage:   "Wire endpoint: tcp://127.0.0.1:<port> (trusted local development)",
			Sources: cli.EnvVars("LAB_RUNTIME_ENDPOINT"),
			Hidden:  hidden,
		},
		&cli.DurationFlag{
			Name:    "runtime-connect-timeout",
			Usage:   "Wire connection and handshake timeout (must be positive)",
			Value:   runtime.DefaultConnectTimeout,
			Sources: cli.EnvVars("LAB_RUNTIME_CONNECT_TIMEOUT"),
			Hidden:  hidden,
		},
	}
}
