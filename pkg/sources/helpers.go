package sources

import (
	"path/filepath"
)

func IsSupportedFile(name string) bool {
	switch filepath.Ext(name) {
	case ".fql", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func ToDir(path string) string {
	if filepath.Ext(path) == "" {
		return path
	}

	return filepath.Dir(path)
}
