package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/go-waitfor/waitfor"
	"github.com/go-waitfor/waitfor-http"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/MontFerret/lab/cdn"
	"github.com/MontFerret/lab/reporters"
	"github.com/MontFerret/lab/runner"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
	"github.com/MontFerret/lab/testing"
)

func toDirectories(values []string) ([]cdn.Directory, error) {
	res := make([]cdn.Directory, 0, len(values))

	for _, entry := range values {
		dir, err := cdn.NewDirectoryFrom(entry)

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

func createCDNManager(dirs []cdn.Directory) (*cdn.Manager, error) {
	m, err := cdn.New()

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
	waitFor := c.StringSlice("wait")

	var locations []string

	if c.NArg() == 0 {
		locations = c.StringSlice("files")
	} else {
		locations = c.Args().Slice()
	}

	if len(locations) == 0 {
		cli.ShowAppHelpAndExit(c, 1)
	}

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

	r, err := runner.New(runner.Options{
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

	stream := r.Run(runner.NewContext(c.Context, params), src)

	return reporters.
		NewConsole(os.Stdout).
		Report(c.Context, stream)
}
