package wirehost

import "github.com/MontFerret/api"

// Params captures portable session settings supplied to a hosted Run callback.
type Params map[string]any

// SetParam retains the last supplied value for a parameter.
func (p Params) SetParam(name string, value any) error {
	p[name] = value

	return nil
}

// SetParams merges maps in session-option order.
func (p Params) SetParams(values map[string]any) error {
	for name, value := range values {
		p[name] = value
	}

	return nil
}

// SetOutputContentType accepts the host's default output codec configuration.
func (p Params) SetOutputContentType(string) error { return nil }

// SetFSRoot accepts the host's default filesystem configuration.
func (p Params) SetFSRoot(string) error { return nil }

var _ api.SessionOptions = Params{}
