package runner

import (
	"path/filepath"
	"strings"
)

func isFQLFile(name string) bool {
	return filepath.Ext(name) == ".fql"
}

func mustFail(name string) bool {
	return strings.HasSuffix(name, ".fail.fql")
}
