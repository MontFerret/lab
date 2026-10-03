package runtime

import (
	"fmt"
	"strings"
)

// FileSystemPolicy configures the built-in runtime's sandboxed filesystem.
// A nil ReadOnly value leaves the runtime's setting unchanged.
type FileSystemPolicy struct {
	Root     string
	ReadOnly *bool
}

func (policy *FileSystemPolicy) hasSettings() bool {
	return policy != nil && (policy.Root != "" || policy.ReadOnly != nil)
}

func (policy *FileSystemPolicy) validate() error {
	if policy == nil || policy.Root == "" {
		return nil
	}

	if strings.TrimSpace(policy.Root) == "" {
		return fmt.Errorf("--policy-fs-root cannot be empty")
	}

	return nil
}
