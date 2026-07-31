package runtime

import (
	"bytes"
	"fmt"
	"os"

	"github.com/MontFerret/ferret/v2/pkg/source"
)

func resolveBinaryScript(query *source.Source) (string, func() error, error) {
	if query == nil {
		return "", nil, fmt.Errorf("binary runtime source cannot be nil")
	}

	content := []byte(query.Content())
	name := query.Name()

	if info, err := os.Stat(name); err == nil && info.Mode().IsRegular() {
		if current, readErr := os.ReadFile(name); readErr == nil && bytes.Equal(current, content) {
			return name, func() error { return nil }, nil
		}
	}

	file, err := os.CreateTemp("", "lab-runtime-*.fql")
	if err != nil {
		return "", nil, fmt.Errorf("create temporary runtime script: %w", err)
	}

	path := file.Name()
	cleanup := func() error {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove temporary runtime script %s: %w", path, err)
		}

		return nil
	}

	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		_ = cleanup()

		return "", nil, fmt.Errorf("write temporary runtime script %s: %w", path, err)
	}

	if err := file.Close(); err != nil {
		_ = cleanup()

		return "", nil, fmt.Errorf("close temporary runtime script %s: %w", path, err)
	}

	return path, cleanup, nil
}
