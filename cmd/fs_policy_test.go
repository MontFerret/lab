package cmd

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"

	"github.com/MontFerret/lab/v2/pkg/runtime"
)

func TestFSPolicyFlags(t *testing.T) {
	policy, err := runFSPolicyCommand(t, "--policy-fs-root= ./fixtures ", "--policy-fs-read-only")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if policy == nil {
		t.Fatal("expected filesystem policy")
	}

	if policy.Root != "./fixtures" {
		t.Fatalf("expected trimmed root, got %q", policy.Root)
	}

	if policy.ReadOnly == nil || !*policy.ReadOnly {
		t.Fatal("expected read-only policy")
	}
}

func TestFSPolicyPreservesExplicitFalse(t *testing.T) {
	policy, err := runFSPolicyCommand(t, "--policy-fs-read-only=false")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if policy == nil || policy.ReadOnly == nil || *policy.ReadOnly {
		t.Fatalf("expected explicit false read-only value, got %#v", policy)
	}
}

func TestFSPolicyFlagsRemainUnsetByDefault(t *testing.T) {
	policy, err := runFSPolicyCommand(t)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if policy != nil {
		t.Fatalf("expected no filesystem policy, got %#v", policy)
	}
}

func TestFSPolicyFlagsRejectExplicitBlankRoot(t *testing.T) {
	_, err := runFSPolicyCommand(t, "--policy-fs-root= \t ")
	if err == nil || err.Error() != "--policy-fs-root cannot be empty" {
		t.Fatalf("expected blank root error, got %v", err)
	}
}

func runFSPolicyCommand(t *testing.T, args ...string) (*runtime.FileSystemPolicy, error) {
	t.Helper()

	var policy *runtime.FileSystemPolicy
	command := &cli.Command{
		Name:  "run",
		Flags: fsPolicyFlags(false),
		Action: func(_ context.Context, cmd *cli.Command) error {
			var err error
			policy, err = fsPolicyFromCommand(cmd)

			return err
		},
	}

	err := command.Run(context.Background(), append([]string{"run"}, args...))

	return policy, err
}
