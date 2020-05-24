package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	sysRuntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	waitfor "github.com/ziflex/waitfor/pkg/runner"

	"github.com/MontFerret/lab/cdn"
	"github.com/MontFerret/lab/reporters"
	"github.com/MontFerret/lab/runner"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

var version string

type Directory struct {
	Name string
	Path string
	Port int
}

func toDirectories(values []string) ([]Directory, error) {
	res := make([]Directory, 0, len(values))

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
	m, err := cdn.New()

	if err != nil {
		return nil, err
	}

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
		Name:        "lab",
		Usage:       "run FQL scripts",
		Description: "Ferret scripts runner",
		Version:     version,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "files",
				Aliases: []string{"f"},
				EnvVars: []string{"FERRET_LAB_FILES"},
				Usage:   "Location of FQL script files to run",
			},
			&cli.StringFlag{
				Name:    "cdp",
				Value:   "http://127.0.0.1:9222",
				Usage:   "Chrome DevTools Protocol address",
				EnvVars: []string{"FERRET_LAB_CDP"},
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
				Usage:   "url to remote Ferret runtime (http, https or bin)",
				EnvVars: []string{"FERRET_LAB_RUNTIME"},
			},
			&cli.StringSliceFlag{
				Name:    "runtime-param",
				Usage:   "params for remote Ferret runtime (--runtime-param=headers:{\"KeyId\": \"abcd\"} --runtime-param=path:\"/ferret\" })",
				EnvVars: []string{"FERRET_LAB_RUNTIME_PARAM"},
			},
			&cli.Uint64Flag{
				Name:    "concurrency",
				Usage:   "number of multiple tests to run at a time",
				EnvVars: []string{"FERRET_LAB_CONCURRENCY"},
				Value:   uint64(sysRuntime.NumCPU()),
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
				Usage:       "tests and waits on the availability of remote resources (--wait http://127.0.0.1:9222/json/version --wait postgres://locahost:5432/mydb)",
				EnvVars:     []string{"FERRET_LAB_WAIT"},
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				TakesFile:   false,
				Value:       nil,
				DefaultText: "",
				HasBeenSet:  false,
			},
			&cli.Uint64Flag{
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
			&cli.Uint64Flag{
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

			// Pass termination down the service tree
			ctx, cancel := context.WithCancel(c.Context)

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			signal.Notify(ch, os.Kill)

			go func() {
				for {
					<-ch
					cancel()
				}
			}()

			defer cancel()

			if len(waitFor) > 0 {
				err := waitfor.Test(
					ctx,
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

			rt, err := runtime.New(runtime.Options{
				RemoteURL: c.String("runtime"),
				CDP:       c.String("cdp"),
				Params:    runtimeParams,
			})

			if err != nil {
				return cli.Exit(err, 1)
			}

			r, err := runner.New(rt, c.Uint64("concurrency"))

			if err != nil {
				return cli.Exit(err, 1)
			}

			var locations []string

			if c.NArg() == 0 {
				locations = c.StringSlice("files")
			} else {
				locations = c.Args().Slice()
			}

			if len(locations) == 0 {
				cli.ShowAppHelpAndExit(c, 1)
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

			err = cdnManager.Start(ctx)

			if err != nil {
				return cli.Exit(errors.Wrap(err, "failed to start local server for CDN"), 1)
			}

			stream := r.Run(runner.NewContext(ctx, params), src)

			err = reporters.
				NewConsole(os.Stdout).
				Report(ctx, stream)

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
