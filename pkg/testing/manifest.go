package testing

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type (
	SuiteManifest struct {
		Timeout uint64              `yaml:"timeout"`
		Query   ScriptManifest      `yaml:"query"`
		Assert  *ScriptManifest     `yaml:"assert"`
		Expect  ExpectationManifest `yaml:"expect"`
	}

	ScriptManifest struct {
		Text   string                 `yaml:"text"`
		Ref    string                 `yaml:"ref"`
		Params map[string]interface{} `yaml:"params"`
	}

	ExpectationManifest struct {
		Error *ErrorExpectationManifest `yaml:"error,omitempty"`
	}

	ErrorExpectationManifest struct {
		Contains string `yaml:"contains,omitempty"`
	}
)

// UnmarshalYAML rejects unknown nested fields without enabling strict decoding
// for the rest of the suite manifest.
func (manifest *ErrorExpectationManifest) UnmarshalYAML(unmarshal func(interface{}) error) error {
	decoded := struct {
		Contains string                 `yaml:"contains,omitempty"`
		Unknown  map[string]interface{} `yaml:",inline"`
	}{}

	if err := unmarshal(&decoded); err != nil {
		return err
	}

	if len(decoded.Unknown) > 0 {
		fields := make([]string, 0, len(decoded.Unknown))

		for field := range decoded.Unknown {
			fields = append(fields, field)
		}

		sort.Strings(fields)

		if len(fields) == 1 {
			return fmt.Errorf("expect.error contains unsupported field %q", fields[0])
		}

		quotedFields := make([]string, len(fields))

		for i, field := range fields {
			quotedFields[i] = fmt.Sprintf("%q", field)
		}

		return fmt.Errorf("expect.error contains unsupported fields %s", strings.Join(quotedFields, ", "))
	}

	manifest.Contains = decoded.Contains

	return nil
}

func (manifest SuiteManifest) validate() error {
	if err := manifest.Query.validate(); err != nil {
		return errors.Wrap(err, "query")
	}

	if manifest.Expect.Error != nil {
		if manifest.Assert != nil {
			return errors.New("expect.error cannot be combined with assert")
		}

		return nil
	}

	if manifest.Assert == nil {
		return errors.New("assert: ref or text must have value")
	}

	if err := manifest.Assert.validate(); err != nil {
		return errors.Wrap(err, "assert")
	}

	return nil
}

func (manifest ScriptManifest) validate() error {
	if manifest.Ref == "" && manifest.Text == "" {
		return errors.New("ref or text must have value")
	}

	if manifest.Ref != "" && manifest.Text != "" {
		return errors.New("only either ref or text must have value")
	}

	return nil
}

func (manifest ScriptManifest) runtimeParams(params Params) map[string]interface{} {
	params.SetUserValues(manifest.Params)

	return params.ToMap()
}

func (expected ErrorExpectationManifest) evaluate(actual error) error {
	if actual == nil {
		return errors.New("expected query to fail, but it completed successfully")
	}

	if expected.Contains != "" && !strings.Contains(actual.Error(), expected.Contains) {
		return fmt.Errorf("expected error containing %q, got: %v", expected.Contains, actual)
	}

	return nil
}
