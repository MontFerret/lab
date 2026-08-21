package cmd

import (
	"fmt"
	"strings"

	"github.com/MontFerret/lab/v2/pkg/testing"
)

func toParamBindings(values []string, paramValues []string) ([]testing.ParamBinding, error) {
	bindings := make([]testing.ParamBinding, 0, len(values))

	for _, value := range values {
		targetValue, sourceValue, found := strings.Cut(value, "=")
		if !found {
			return nil, fmt.Errorf("invalid parameter binding %q: expected <target>=@<source>", value)
		}

		targetValue = strings.TrimSpace(targetValue)
		sourceValue = strings.TrimSpace(sourceValue)

		if strings.HasPrefix(targetValue, "@") {
			return nil, fmt.Errorf("invalid parameter binding target %q: must not start with @", targetValue)
		}

		target, err := testing.ParseParamPath(targetValue)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter binding target %q: %w", targetValue, err)
		}

		if target.String() == "lab" || strings.HasPrefix(target.String(), "lab.") {
			return nil, fmt.Errorf("invalid parameter binding target %q: \"lab\" is reserved", targetValue)
		}

		if !strings.HasPrefix(sourceValue, "@") {
			return nil, fmt.Errorf("invalid parameter binding source %q for target %q: must start with @", sourceValue, targetValue)
		}

		source, err := testing.ParseParamPath(strings.TrimPrefix(sourceValue, "@"))
		if err != nil {
			return nil, fmt.Errorf("invalid parameter binding source %q for target %q: %w", sourceValue, targetValue, err)
		}

		for _, existing := range bindings {
			if target.String() == existing.Target.String() {
				return nil, fmt.Errorf("duplicate parameter binding target %q", target.String())
			}

			if target.Overlaps(existing.Target) {
				return nil, fmt.Errorf("parameter binding target %q overlaps target %q", target.String(), existing.Target.String())
			}
		}

		bindings = append(bindings, testing.ParamBinding{
			Target: target,
			Source: source,
		})
	}

	for _, value := range paramValues {
		pair := strings.SplitN(value, ":", 2)
		if len(pair) < 2 {
			continue
		}

		target, err := testing.ParseParamPath(pair[0])
		if err != nil {
			continue
		}

		for _, binding := range bindings {
			if target.String() == binding.Target.String() {
				return nil, fmt.Errorf("parameter target %q is assigned by both --param and --param-bind", target.String())
			}

			if target.Overlaps(binding.Target) {
				return nil, fmt.Errorf("parameter binding target %q overlaps --param target %q", binding.Target.String(), target.String())
			}
		}
	}

	return bindings, nil
}
