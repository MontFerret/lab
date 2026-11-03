package cmd

import (
	"errors"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/runtime"
)

func toRuntimeOptions(values []string) (map[string]any, error) {
	var entries []string
	for _, value := range values {
		start, depth := 0, 0
		quoted, escaped := false, false
		for index, char := range value {
			if quoted {
				if escaped {
					escaped = false
				} else if char == '\\' {
					escaped = true
				} else if char == '"' {
					quoted = false
				}

				continue
			}

			switch char {
			case '"':
				quoted = true
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			case ',':
				if depth == 0 {
					entries = append(entries, strings.TrimSpace(value[start:index]))
					start = index + 1
				}
			}
		}

		entries = append(entries, strings.TrimSpace(value[start:]))
	}

	options, err := toParams(entries)
	if err != nil {
		// The shared parameter decoder includes input text in some errors.
		return nil, errors.New("invalid --runtime-option: expected name:<JSON value> (strings require JSON quotes)")
	}

	return options, nil
}

func runtimeOptionsFromCommand(cmd *cli.Command) (runtime.Options, error) {
	options := runtime.Options{Target: cmd.String("runtime")}

	fsPolicy, err := fsPolicyFromCommand(cmd)
	if err != nil {
		return options, err
	}

	httpPolicy, err := httpPolicyFromCommand(cmd)
	if err != nil {
		return options, err
	}

	values, _ := cmd.Value("runtime-option").([]string)
	adapterOptions, err := toRuntimeOptions(values)
	if err != nil {
		return options, err
	}

	options.AdapterOptions = adapterOptions
	options.FSPolicy, options.HTTPPolicy = fsPolicy, httpPolicy

	return options, options.Validate()
}
