package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	ferretrt "github.com/MontFerret/ferret/v2/pkg/runtime"
	cdn2 "github.com/MontFerret/lab/v2/pkg/cdn"
	"github.com/MontFerret/lab/v2/pkg/runtime"
)

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
			return nil, fmt.Errorf("failed to parse JSON for param %q: %w", key, err)
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
		CDPAddress: cdpAddressFromContext(c),
		Params:     params,
	})

	if err != nil {
		return nil, err
	}

	return rt, nil
}

func cdpAddressFromContext(c *cli.Context) string {
	if c != nil {
		if address := c.String("cdp"); address != "" {
			return address
		}
	}

	return defaultCDPAddress
}

func locationsFromContext(c *cli.Context) ([]string, bool) {
	if c.NArg() == 0 {
		locations := c.StringSlice("files")

		return locations, len(locations) > 0
	}

	locations := c.Args().Slice()

	return locations, len(locations) > 0
}
