package cmd

import "github.com/urfave/cli/v2"

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
			EnvVars: []string{"LAB_FILES"},
			Usage:   "location of FQL script files to run",
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Usage:   "test timeout in seconds",
			EnvVars: []string{"LAB_TIMEOUT"},
			Value:   30,
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "cdp",
			Value:   "http://127.0.0.1:9222",
			Usage:   "Chrome DevTools Protocol address",
			EnvVars: []string{"LAB_CDP"},
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "reporter",
			Usage:   "reporter (console, simple)",
			EnvVars: []string{"LAB_REPORTER"},
			Value:   "console",
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    "runtime",
			Aliases: []string{"r"},
			Usage:   "url to remote Ferret runtime (http, https or bin)",
			EnvVars: []string{"LAB_RUNTIME"},
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "runtime-param",
			Aliases: []string{"rp"},
			Usage:   "params for remote Ferret runtime (--runtime-param=headers:{\"KeyId\": \"abcd\"} --runtime-param=path:\"/ferret\" })",
			EnvVars: []string{"LAB_RUNTIME_PARAM"},
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "concurrency",
			Aliases: []string{"c"},
			Usage:   "number of multiple tests to run at a time",
			EnvVars: []string{"LAB_CONCURRENCY"},
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "times",
			Usage:   "number of times to run each test",
			EnvVars: []string{"LAB_TIMES"},
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "attempts",
			Aliases: []string{"a"},
			Usage:   "number of times to re-run failed tests",
			EnvVars: []string{"LAB_ATTEMPTS"},
			Value:   1,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "times-interval",
			Usage:   "interval between test cycles in seconds",
			EnvVars: []string{"LAB_TIMES_INTERVAL"},
			Value:   0,
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "cdn",
			Usage:   "file or directory to serve via HTTP (./dir as default or ./dir@name with alias)",
			EnvVars: []string{"LAB_CDN"},
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "param",
			Aliases: []string{"p"},
			Usage:   "query parameter (--param=foo:\"bar\", --param=id:1)",
			EnvVars: []string{"LAB_PARAM"},
			Hidden:  hidden,
		},
		&cli.StringSliceFlag{
			Name:    "wait",
			Aliases: []string{"w"},
			Usage:   "tests and waits on the availability of remote resources (--wait http://127.0.0.1:9222/json/version --wait postgres://locahost:5432/mydb)",
			EnvVars: []string{"LAB_WAIT"},
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "wait-timeout",
			Aliases: []string{"wt"},
			Usage:   "wait timeout in seconds",
			EnvVars: []string{"LAB_WAIT_TIMEOUT"},
			Value:   5,
			Hidden:  hidden,
		},
		&cli.Uint64Flag{
			Name:    "wait-attempts",
			Usage:   "wait attempts",
			EnvVars: []string{"LAB_WAIT_ATTEMPTS"},
			Value:   5,
			Hidden:  hidden,
		},
	}
}
