package main

import (
	"encoding/json"
	"fmt"
	"os"
	sysRuntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/MontFerret/lab/cdn"
	"github.com/MontFerret/lab/reporters"
	"github.com/MontFerret/lab/runner"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type Directory struct {
	Name string
	Path string
	Port int
}

func toDirectories(values []string) ([]Directory, error) {
	res := make([]Directory, len(values))

	for _, entry := range values {
		dir := Directory{}
		pathAndPort := strings.Split(entry, ":")

		if len(pathAndPort) != 2 {
			return nil, errors.New("invalid directory binding format")
		}

		name := "default"
		path := pathAndPort[0]
		port := pathAndPort[1]

		portAndName := strings.Split(pathAndPort[1], "@")

		if len(portAndName) == 2 {
			port = portAndName[0]
			name = portAndName[1]
		}

		portInt, err := strconv.Atoi(port)

		if err != nil {
			return nil, err
		}

		dir.Name = name
		dir.Path = path
		dir.Port = portInt

		res = append(res, dir)
	}

	return res, nil
}

func toParams(values []string) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	for _, entry := range values {
		pair := strings.SplitN(entry, ":", 2)

		if len(pair) < 2 {
			return nil, core.Error(core.ErrInvalidArgument, entry)
		}

		var value interface{}
		key := pair[0]

		err := json.Unmarshal([]byte(pair[1]), &value)

		if err != nil {
			fmt.Println(pair[1])
			return nil, err
		}

		res[key] = value
	}

	return res, nil
}

func createCDNManager(dirs []Directory) (*cdn.Manager, error) {
	m := cdn.New()

	for _, dir := range dirs {
		err := m.Add(cdn.NewNode(cdn.NodeSettings{
			Name:   dir.Name,
			Port:   dir.Port,
			Dir:    dir.Path,
			Prefix: "",
		}))

		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func waitBeforeStart(deps []string, timeout time.Duration, attempts int) error {
	for _, url := range deps {
		client := retryablehttp.NewClient()
		client.RetryWaitMax = time.Duration(timeout) * time.Second
		client.RetryMax = attempts
		client.Logger = nil

		if _, err := client.Get(url); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "lab",
		Usage: "run FQL scripts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "cdp",
				Value:   "http://127.0.0.1:9222",
				Usage:   "Chrome DevTools Protocol address",
				EnvVars: []string{"FERRET_LAB_CDP"},
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "filter test files",
				Value:   "",
				EnvVars: []string{"FERRET_LAB_FILTER"},
			},
			&cli.StringFlag{
				Name:    "reporter",
				Aliases: []string{"r"},
				Usage:   "reporter (console, simple)",
				EnvVars: []string{"FERRET_LAB_REPORTER"},
				Value:   "console",
			},
			&cli.StringFlag{
				Name:    "runtime",
				Usage:   "url to remote Ferret runtime",
				EnvVars: []string{"FERRET_LAB_RUNTIME"},
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Usage:   "number of multiple tests to run at a time",
				EnvVars: []string{"FERRET_LAB_CONCURRENCY"},
				Value:   sysRuntime.NumCPU() * 2,
			},
			&cli.StringSliceFlag{
				Name:        "dir",
				Aliases:     []string{"d"},
				Usage:       "file or directory to serve (./dir:8080 as default or ./dir:8080@name as named)",
				EnvVars:     []string{"FERRET_LAB_DIR"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				TakesFile:   false,
				Value:       cli.NewStringSlice(),
				DefaultText: "",
				HasBeenSet:  false,
			},
			&cli.StringSliceFlag{
				Name:        "param",
				Aliases:     []string{"p"},
				Usage:       "query parameter (--param=foo:\"bar\", --param=id:1)",
				EnvVars:     []string{"FERRET_LAB_PARAM"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				TakesFile:   false,
				Value:       nil,
				DefaultText: "",
				HasBeenSet:  false,
			},
			&cli.StringSliceFlag{
				Name:        "wait",
				Aliases:     []string{"w"},
				Usage:       "waits for a 3rd party service by calling its endpoint (--wait http://127.0.0.1:9222/json/version)",
				EnvVars:     []string{"FERRET_LAB_WAIT"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				TakesFile:   false,
				Value:       nil,
				DefaultText: "",
				HasBeenSet:  false,
			},
			&cli.IntFlag{
				Name:        "wait-timeout",
				Aliases:     nil,
				Usage:       "wait timeout in seconds",
				EnvVars:     []string{"FERRET_LAB_WAIT_TIMEOUT"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				Value:       5,
				DefaultText: "",
				Destination: nil,
				HasBeenSet:  false,
			},
			&cli.IntFlag{
				Name:        "wait-attempts",
				Aliases:     nil,
				Usage:       "wait attempts",
				EnvVars:     []string{"FERRET_LAB_WAIT_TRY"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				Value:       5,
				DefaultText: "",
				Destination: nil,
				HasBeenSet:  false,
			},
		},
		Action: func(c *cli.Context) error {
			waitFor := c.StringSlice("wait")

			if len(waitFor) > 0 {
				if err := waitBeforeStart(waitFor, time.Duration(c.Int("wait-timeout")), c.Int("wait-attempts")); err != nil {
					return cli.Exit(errors.Wrap(err, "timeout"), 1)
				}
			}

			rt := runtime.New(runtime.Options{
				RemoteURL: c.String("runtime"),
				CDP:       c.String("cdp"),
			})
			r, err := runner.New(rt, c.Int("concurrency"))

			if err != nil {
				return cli.Exit(err, 1)
			}

			var locations []string

			if c.NArg() == 0 {
				wd, err := os.Getwd()

				if err != nil {
					return cli.Exit(err, 1)
				}

				locations = []string{wd}
			} else {
				locations = c.Args().Slice()
			}

			src, err := sources.New(locations...)

			if err != nil {
				return cli.Exit(err, 1)
			}

			params := runner.NewParams()

			userParams, err := toParams(c.StringSlice("p"))

			if err != nil {
				return cli.Exit(err, 1)
			}

			params.SetUserValues(userParams)

			dirs, err := toDirectories(c.StringSlice("d"))

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

			stream := r.Run(runner.NewContext(c.Context, params), src)

			err = reporters.
				NewConsole(os.Stdout).
				Report(c.Context, stream)

			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		fmt.Println("failed to start the app")

		os.Exit(1)
	}
}
