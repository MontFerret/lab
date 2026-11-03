package apiruntime

import (
	"errors"

	"github.com/MontFerret/api"
)

type (
	// Params captures parameter settings supplied to a Run callback.
	Params map[string]any

	// Options captures all portable session settings.
	Options struct {
		Params
		FSRoot      string
		ContentType string
	}
)

// NewParams evaluates non-nil callbacks once in order, joining their errors.
func NewParams(opts ...api.SessionOption) (Params, error) {
	params := Params{}
	var failures []error
	for _, option := range opts {
		if option != nil {
			failures = append(failures, option(params))
		}
	}

	return params, errors.Join(failures...)
}

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

// SetFSRoot captures the portable filesystem setting.
func (o *Options) SetFSRoot(root string) error {
	o.FSRoot = root

	return nil
}

// SetOutputContentType captures the portable output setting.
func (o *Options) SetOutputContentType(contentType string) error {
	o.ContentType = contentType

	return nil
}

var _ api.SessionOptions = (*Options)(nil)
