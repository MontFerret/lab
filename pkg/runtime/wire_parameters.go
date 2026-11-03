package runtime

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// YAML suite objects use interface-keyed maps. Convert their string keys to
// Wire's portable object shape without JSON round trips or mutating inputs.
func wireParameters(params map[string]any) (map[string]any, error) {
	return wireParameterMap(params, 0)
}

func wireParameterMap(values map[string]any, depth int) (map[string]any, error) {
	if values == nil {
		return nil, nil
	}

	converted := make(map[string]any, len(values))
	for _, key := range slices.Sorted(maps.Keys(values)) {
		portable, err := wireParameter(values[key], depth)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", key, err)
		}

		converted[key] = portable
	}

	return converted, nil
}

func wireParameter(value any, depth int) (any, error) {
	// Match the released Wire codec's depth bound, including cyclic Go values.
	if depth >= 64 {
		return nil, errors.New("value nesting exceeds 64 levels")
	}

	switch value := value.(type) {
	case map[string]any:
		return wireParameterMap(value, depth+1)
	case map[any]any:
		converted := make(map[string]any, len(value))
		names := make([]string, 0, len(value))
		for key := range value {
			name, ok := key.(string)
			if !ok {
				return nil, errors.New("wire parameter objects require string keys")
			}

			names = append(names, name)
		}

		slices.Sort(names)
		for _, name := range names {
			portable, err := wireParameter(value[name], depth+1)
			if err != nil {
				return nil, fmt.Errorf("object field %q: %w", name, err)
			}

			converted[name] = portable
		}

		return converted, nil
	case []any:
		if value == nil {
			return value, nil
		}

		converted := make([]any, len(value))
		for index, item := range value {
			portable, err := wireParameter(item, depth+1)
			if err != nil {
				return nil, fmt.Errorf("array item %d: %w", index, err)
			}

			converted[index] = portable
		}

		return converted, nil
	default:
		return value, nil
	}
}
