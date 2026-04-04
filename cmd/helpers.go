package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	ferretrt "github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/staticserver"
)

func toServeEntries(values []string) (staticserver.ServeEntries, error) {
	return staticserver.ParseServeEntries(values)
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

func createStaticServerManagerFromCommand(cmd *cli.Command, entries staticserver.ServeEntries) (*staticserver.Manager, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	manager, err := staticserver.NewManager(staticServerSettingsFromCommand(cmd))
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if err := manager.Bind(entry); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

func staticServerSettingsFromCommand(cmd *cli.Command) staticserver.Settings {
	if cmd == nil {
		return staticserver.Settings{}
	}

	return staticserver.Settings{
		BindHost:      cmd.String("serve-bind"),
		AdvertiseHost: cmd.String("serve-host"),
	}
}

func newRuntime(cmd *cli.Command, params map[string]interface{}) (runtime.Runtime, error) {
	rt, err := runtime.New(runtime.Options{
		Type:       cmd.String("runtime"),
		CDPAddress: cdpAddressFromCommand(cmd),
		Params:     params,
	})

	if err != nil {
		return nil, err
	}

	return rt, nil
}

func cdpAddressFromCommand(cmd *cli.Command) string {
	if cmd != nil {
		if address := cmd.String("cdp"); address != "" {
			return address
		}
	}

	return defaultCDPAddress
}

func locationsFromCommand(cmd *cli.Command) ([]string, bool) {
	if cmd.NArg() == 0 {
		locations := cmd.StringSlice("files")

		return locations, len(locations) > 0
	}

	locations := cmd.Args().Slice()

	return locations, len(locations) > 0
}
