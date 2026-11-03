package cmd

import (
	"strings"

	"github.com/urfave/cli/v3"
)

// Collect each occurrence intact; the CLI's slice splitter cannot distinguish
// option separators from commas inside JSON strings and objects.
type runtimeOptionValues struct {
	values *[]string
}

func (runtimeOptionValues) Create(value []string, destination *[]string, _ cli.NoConfig) cli.Value {
	*destination = append([]string(nil), value...)

	return &runtimeOptionValues{values: destination}
}

func (runtimeOptionValues) ToString(values []string) string {
	return strings.Join(values, ", ")
}

func (v *runtimeOptionValues) Set(value string) error {
	*v.values = append(*v.values, value)

	return nil
}

func (v *runtimeOptionValues) Get() any {
	return *v.values
}

func (v *runtimeOptionValues) String() string {
	return v.ToString(*v.values)
}
