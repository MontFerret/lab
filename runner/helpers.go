package runner

import (
	"strings"
)

func mustFail(name string) bool {
	return strings.HasSuffix(name, ".fail.fql")
}
