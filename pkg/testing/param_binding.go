package testing

import (
	"fmt"
	"regexp"
	"strings"
)

var paramPathSegmentExp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

type (
	// ParamPath is a validated, case-sensitive path through materialized parameters.
	ParamPath struct {
		segments []string
	}

	// ParamBinding maps a user parameter target to an existing parameter source.
	ParamBinding struct {
		Target ParamPath
		Source ParamPath
	}
)

// ParseParamPath validates and constructs a dot-separated parameter path.
func ParseParamPath(value string) (ParamPath, error) {
	if value == "" {
		return ParamPath{}, fmt.Errorf("parameter path cannot be empty")
	}

	segments := strings.Split(value, ".")
	for _, segment := range segments {
		if !paramPathSegmentExp.MatchString(segment) {
			return ParamPath{}, fmt.Errorf("invalid parameter path %q", value)
		}
	}

	return ParamPath{segments: segments}, nil
}

// String returns the canonical dot-separated path.
func (path ParamPath) String() string {
	return strings.Join(path.segments, ".")
}

// Overlaps reports whether either path is the other path or its ancestor.
func (path ParamPath) Overlaps(other ParamPath) bool {
	if len(path.segments) == 0 || len(other.segments) == 0 {
		return false
	}

	limit := min(len(path.segments), len(other.segments))
	for i := 0; i < limit; i++ {
		if path.segments[i] != other.segments[i] {
			return false
		}
	}

	return true
}

func (path ParamPath) lookup(values map[string]any) (any, bool) {
	current := values

	for i, segment := range path.segments {
		value, exists := current[segment]
		if !exists {
			return nil, false
		}

		if i == len(path.segments)-1 {
			return value, true
		}

		next, ok := value.(map[string]any)
		if !ok {
			return nil, false
		}

		current = next
	}

	return nil, false
}

func (path ParamPath) set(values map[string]any, value any) error {
	current := values

	for i, segment := range path.segments {
		if i == len(path.segments)-1 {
			if _, exists := current[segment]; exists {
				return fmt.Errorf("parameter target %q already exists", path.String())
			}

			current[segment] = value

			return nil
		}

		nextValue, exists := current[segment]
		if !exists {
			next := make(map[string]any)
			current[segment] = next
			current = next

			continue
		}

		next, ok := nextValue.(map[string]any)
		if !ok {
			return fmt.Errorf("parameter target %q traverses non-object path %q", path.String(), strings.Join(path.segments[:i+1], "."))
		}

		current = next
	}

	return nil
}
