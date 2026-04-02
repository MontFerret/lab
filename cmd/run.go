package cmd

import "github.com/urfave/cli/v3"

const defaultCDPAddress = "http://127.0.0.1:9222"

func RunCommand() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run FQL test scripts",
		UsageText: "lab run [options] [files...]",
		Flags:     RunFlags(false),
		Action:    RunAction,
	}
}

func RunFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "files",
			Aliases: []string{"f"},
			Sources: cli.EnvVars("LAB_FILES"),
			Usage:   "location of FQL script files to run",
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Usage:   "test timeout in seconds",
			Sources: cli.EnvVars("LAB_TIMEOUT"),
			Value:   30,
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "cdp",
			Value:   defaultCDPAddress,
			Usage:   "Chrome DevTools Protocol address",
			Sources: cli.EnvVars("LAB_CDP"),
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "reporter",
			Usage:   "reporter (console, simple)",
			Sources: cli.EnvVars("LAB_REPORTER"),
			Value:   "console",
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "runtime",
			Aliases: []string{"r"},
			Usage:   "url to remote Ferret runtime (http, https or bin)",
			Sources: cli.EnvVars("LAB_RUNTIME"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "runtime-param",
			Aliases: []string{"rp"},
			Usage:   "params for remote Ferret runtime (--runtime-param=headers:{\"KeyId\": \"abcd\"} --runtime-param=path:\"/ferret\")",
			Sources: cli.EnvVars("LAB_RUNTIME_PARAM"),
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "concurrency",
			Aliases: []string{"c"},
			Usage:   "number of multiple tests to run at a time",
			Sources: cli.EnvVars("LAB_CONCURRENCY"),
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "times",
			Usage:   "number of times to run each test",
			Sources: cli.EnvVars("LAB_TIMES"),
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "attempts",
			Aliases: []string{"a"},
			Usage:   "number of times to re-run failed tests",
			Sources: cli.EnvVars("LAB_ATTEMPTS"),
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "times-interval",
			Usage:   "interval between test cycles in seconds",
			Sources: cli.EnvVars("LAB_TIMES_INTERVAL"),
			Value:   0,
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "cdn",
			Usage:   "file or directory to serve via HTTP (./dir as default or ./dir@name with alias)",
			Sources: cli.EnvVars("LAB_CDN"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "param",
			Aliases: []string{"p"},
			Usage:   "query parameter (--param=foo:\"bar\", --param=id:1)",
			Sources: cli.EnvVars("LAB_PARAM"),
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "wait",
			Aliases: []string{"w"},
			Usage:   "tests and waits on the availability of remote resources (--wait http://127.0.0.1:9222/json/version --wait postgres://localhost:5432/mydb)",
			Sources: cli.EnvVars("LAB_WAIT"),
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "wait-timeout",
			Aliases: []string{"wt"},
			Usage:   "wait timeout in seconds",
			Sources: cli.EnvVars("LAB_WAIT_TIMEOUT"),
			Value:   5,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "wait-attempts",
			Usage:   "wait attempts",
			Sources: cli.EnvVars("LAB_WAIT_ATTEMPTS"),
			Value:   5,
			Hidden:  hidden,
		},
	}
}
