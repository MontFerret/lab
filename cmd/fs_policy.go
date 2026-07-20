package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/runtime"
)

func fsPolicyFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "policy-fs-root",
			Usage:   "filesystem root directory for the built-in runtime",
			Sources: cli.EnvVars("LAB_POLICY_FS_ROOT"),
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    "policy-fs-read-only",
			Usage:   "make the built-in runtime filesystem read-only",
			Sources: cli.EnvVars("LAB_POLICY_FS_READ_ONLY"),
			Hidden:  hidden,
		},
	}
}

func fsPolicyFromCommand(cmd *cli.Command) (*runtime.FileSystemPolicy, error) {
	if cmd == nil {
		return nil, nil
	}

	rootSet := cmd.IsSet("policy-fs-root")
	readOnlySet := cmd.IsSet("policy-fs-read-only")
	if !rootSet && !readOnlySet {
		return nil, nil
	}

	policy := &runtime.FileSystemPolicy{
		ReadOnly: cmd.Bool("policy-fs-read-only"),
	}

	if rootSet {
		policy.Root = strings.TrimSpace(cmd.String("policy-fs-root"))
		if policy.Root == "" {
			return nil, fmt.Errorf("--policy-fs-root cannot be empty")
		}
	}

	return policy, nil
}
