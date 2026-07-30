package runtime

import (
	"fmt"
	"strings"
)

// FileSystemPolicy configures the sandboxed filesystem used by built-in and
// binary Ferret runtimes. A nil ReadOnly value leaves the runtime's setting
// unchanged.
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

func (policy *FileSystemPolicy) ferretCLIArgs() []string {
	if policy == nil {
		return nil
	}

	args := make([]string, 0, 2)
	if policy.Root != "" {
		args = append(args, "--policy-fs-root="+policy.Root)
	}

	if policy.ReadOnly != nil {
		args = append(args, fmt.Sprintf("--policy-fs-read-only=%t", *policy.ReadOnly))
	}

	return args
}

func (policy *FileSystemPolicy) conflictingRawFlags() map[string]struct{} {
	flags := make(map[string]struct{}, 2)
	if policy == nil {
		return flags
	}

	if policy.Root != "" {
		flags["--policy-fs-root"] = struct{}{}
	}

	if policy.ReadOnly != nil {
		flags["--policy-fs-read-only"] = struct{}{}
	}

	return flags
}
