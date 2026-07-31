package runtime

import (
	"fmt"
	"net/url"
)

func newConfiguredBuiltin(params map[string]any, fsPolicy *FileSystemPolicy, httpPolicy *HTTPPolicy) (*Builtin, error) {
	if err := fsPolicy.validate(); err != nil {
		return nil, err
	}

	options, err := httpPolicy.validatedFerretOptions()
	if err != nil {
		return nil, fmt.Errorf("HTTP policy: %w", err)
	}

	return newBuiltin(params, fsPolicy, options...)
}

func binaryPath(u *url.URL) string {
	if u.Opaque != "" {
		return u.Opaque
	}

	return u.Host + u.Path
}
