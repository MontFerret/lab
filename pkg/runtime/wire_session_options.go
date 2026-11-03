package runtime

import (
	"fmt"

	"github.com/MontFerret/api"
)

// wireSessionOptions converts YAML objects at parameter setters while preserving
// the remote runtime's other session capabilities and callback evaluation.
type wireSessionOptions struct {
	api.SessionOptions
}

func (o *wireSessionOptions) SetParam(name string, value any) error {
	converted, err := wireParameter(value, 0)
	if err != nil {
		return fmt.Errorf("parameter %q: %w", name, err)
	}

	return o.SessionOptions.SetParam(name, converted)
}

func (o *wireSessionOptions) SetParams(values map[string]any) error {
	converted, err := wireParameters(values)
	if err != nil {
		return err
	}

	return o.SessionOptions.SetParams(converted)
}
