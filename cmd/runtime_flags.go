package cmd

import "github.com/urfave/cli/v3"

func runtimeFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "runtime",
			Aliases: []string{"r"},
			Usage:   "Ferret runtime target (builtin, HTTP/HTTPS URL, or wire+tcp://127.0.0.1:<port>)",
			Sources: cli.EnvVars("LAB_RUNTIME"),
			Hidden:  hidden,
		},
		&cli.FlagBase[[]string, cli.NoConfig, runtimeOptionValues]{
			Name:    "runtime-option",
			Usage:   `adapter option name:<JSON value> (repeatable; HTTP: headers, cookies, path; Wire: connect-timeout, default "5s")`,
			Sources: cli.EnvVars("LAB_RUNTIME_OPTION"),
			Hidden:  hidden,
		},
	}
}
