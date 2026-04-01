package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-waitfor/waitfor"
	http "github.com/go-waitfor/waitfor-http"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	cdn2 "github.com/MontFerret/lab/v2/pkg/cdn"
	"github.com/MontFerret/lab/v2/pkg/reporters"
	runner2 "github.com/MontFerret/lab/v2/pkg/runner"
	"github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	"github.com/MontFerret/lab/v2/pkg/testing"

	ferretrt "github.com/MontFerret/ferret/v2/pkg/runtime"
)

const deprecatedRunWarning = "Warning: bare script execution is deprecated; use `lab run ...` instead."

func toDirectories(values []string) ([]cdn2.Directory, error) {
	res := make([]cdn2.Directory, 0, len(values))

	for _, entry := range values {
		dir, err := cdn2.NewDirectoryFrom(entry)

		if err != nil {
			return nil, err
		}

		res = append(res, dir)
	}

	return res, nil
}

func toParams(values []string) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	for _, entry := range values {
		pair := strings.SplitN(entry, ":", 2)

		if len(pair) < 2 {
			return nil, ferretrt.Error(ferretrt.ErrInvalidArgument, entry)
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

func createCDNManager(dirs []cdn2.Directory) (*cdn2.Manager, error) {
	m, err := cdn2.New()

	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		err := m.Bind(dir)

		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func newRuntime(c *cli.Context, params map[string]interface{}) (runtime.Runtime, error) {
	rt, err := runtime.New(runtime.Options{
		Type:       c.String("runtime"),
		CDPAddress: c.String("cdp"),
		Params:     params,
	})

	if err != nil {
		return nil, err
	}

	return rt, nil
}

func DefaultCommand(c *cli.Context) error {
	locations, ok := locationsFromContext(c)

	if !ok {
		if err := cli.ShowAppHelp(c); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	fmt.Fprintln(appErrWriter(c), deprecatedRunWarning)

	return runScripts(c, locations)
}

func RunAction(c *cli.Context) error {
	locations, ok := locationsFromContext(c)

	if !ok {
		if err := cli.ShowCommandHelp(c, "run"); err != nil {
			return err
		}

		return cli.Exit("", 1)
	}

	return runScripts(c, locations)
}

func locationsFromContext(c *cli.Context) ([]string, bool) {
	if c.NArg() == 0 {
		locations := c.StringSlice("files")

		return locations, len(locations) > 0
	}

	locations := c.Args().Slice()

	return locations, len(locations) > 0
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

func appWriter(c *cli.Context) io.Writer {
	if c != nil && c.App != nil && c.App.Writer != nil {
		return c.App.Writer
	}

	return os.Stdout
}

func appErrWriter(c *cli.Context) io.Writer {
	if c != nil && c.App != nil && c.App.ErrWriter != nil {
		return c.App.ErrWriter
	}

	return os.Stderr
}
