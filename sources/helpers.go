package sources

import "path/filepath"

func isFQLFile(name string) bool {
	return filepath.Ext(name) == ".fql"
}
