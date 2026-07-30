package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

func appendStringSliceFlag(args []string, name string, values []string) []string {
	if values == nil {
		return args
	}

	return append(args, "--"+name+"="+strings.Join(values, ","))
}

func appendBoolFlag(args []string, name string, value *bool) []string {
	if value == nil {
		return args
	}

	return append(args, "--"+name+"="+strconv.FormatBool(*value))
}

func addManagedSliceFlag(flags map[string]struct{}, name string, values []string) {
	if values != nil {
		flags[name] = struct{}{}
	}
}

func addManagedBoolFlag(flags map[string]struct{}, name string, value *bool) {
	if value != nil {
		flags[name] = struct{}{}
	}
}

func validateRawBinaryFlags(rawFlags []string, conflictingFlags map[string]struct{}) error {
	for _, arg := range rawFlags {
		name, _, _ := strings.Cut(arg, "=")

		if _, exists := conflictingFlags[name]; exists {
			return fmt.Errorf("raw binary flag %q conflicts with managed policy flag %q", arg, name)
		}
	}

	return nil
}
