package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/source"
)

type (
	Binary struct {
		path     string
		baseArgs []string
	}

	// BinaryOptions configures a Ferret CLI v2 binary runtime.
	BinaryOptions struct {
		Path   string
		Params map[string]any
		Flags  []string
		// FSPolicy is serialized as Ferret CLI filesystem policy flags.
		FSPolicy *FileSystemPolicy
		// HTTPPolicy is serialized as Ferret CLI HTTP policy flags.
		HTTPPolicy *HTTPPolicy
	}
)

// NewBinary creates an adapter for a Ferret CLI v2 executable.
func NewBinary(opts BinaryOptions) (*Binary, error) {
	if strings.TrimSpace(opts.Path) == "" {
		return nil, errors.New("binary runtime path cannot be empty")
	}

	if err := opts.FSPolicy.validate(); err != nil {
		return nil, fmt.Errorf("filesystem policy: %w", err)
	}

	if err := opts.HTTPPolicy.validate(); err != nil {
		return nil, fmt.Errorf("HTTP policy: %w", err)
	}

	conflictingFlags := opts.FSPolicy.conflictingRawFlags()
	for flag := range opts.HTTPPolicy.conflictingRawFlags() {
		conflictingFlags[flag] = struct{}{}
	}

	if err := validateRawBinaryFlags(opts.Flags, conflictingFlags); err != nil {
		return nil, err
	}

	fsArgs := opts.FSPolicy.ferretCLIArgs()
	httpArgs, err := opts.HTTPPolicy.ferretCLIArgs()

	if err != nil {
		return nil, err
	}

	rt := &Binary{path: opts.Path}
	sharedArgs, err := rt.paramsToArg(opts.Params)
	if err != nil {
		return nil, err
	}

	rt.baseArgs = slices.Concat([]string{"run"}, opts.Flags, fsArgs, httpArgs, sharedArgs)

	return rt, nil
}

func (rt *Binary) Version(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, rt.path, "version")

	out, err := cmd.CombinedOutput()

	if err != nil {
		if len(out) != 0 {
			return "", errors.New(string(out))
		}

		return "", err
	}

	return strings.ReplaceAll(string(out), "\n", ""), nil
}

func (rt *Binary) Run(ctx context.Context, query *source.Source, params map[string]any) ([]byte, error) {
	args, err := rt.runArgs(params)

	if err != nil {
		return nil, err
	}

	var q bytes.Buffer
	q.WriteString(query.Content())

	cmd := exec.CommandContext(ctx, rt.path, args...)
	cmd.Stdin = &q

	out, err := cmd.CombinedOutput()

	if err != nil {
		if len(out) != 0 {
			return nil, errors.New(string(out))
		}

		return nil, err
	}

	return out, nil
}

func (rt *Binary) Close() error {
	return nil
}

func (rt *Binary) paramsToArg(params map[string]any) ([]string, error) {
	args := make([]string, 0, len(params))
	keys := make([]string, 0, len(params))

	for key := range params {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := params[k]
		j, err := json.Marshal(v)

		if err != nil {
			return nil, fmt.Errorf("failed to serialize parameter: %s: %w", k, err)
		}

		args = append(args, fmt.Sprintf("--param=%s=%s", k, j))
	}

	return args, nil
}

func (rt *Binary) runArgs(params map[string]any) ([]string, error) {
	queryArgs, err := rt.paramsToArg(params)
	if err != nil {
		return nil, err
	}

	return slices.Concat(rt.baseArgs, queryArgs), nil
}
