package runtime

import (
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
		protocol Protocol
		path     string
		baseArgs []string
	}

	// BinaryOptions configures a binary runtime.
	BinaryOptions struct {
		Path       string
		Protocol   Protocol
		Params     map[string]any
		Flags      []string
		FSPolicy   *FileSystemPolicy
		HTTPPolicy *HTTPPolicy
	}
)

// NewBinary creates an adapter for a Ferret-compatible binary executable.
func NewBinary(opts BinaryOptions) (*Binary, error) {
	if strings.TrimSpace(opts.Path) == "" {
		return nil, errors.New("binary runtime path cannot be empty")
	}

	protocol := opts.Protocol

	if err := validateBinaryRuntimeProtocol(protocol); err != nil {
		return nil, err
	}

	if protocol == ProtocolCLIDirect {
		if err := validateCLIDirectProtocol(opts); err != nil {
			return nil, err
		}

		return &Binary{
			path:     opts.Path,
			protocol: protocol,
		}, nil
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

	rt := &Binary{
		path:     opts.Path,
		protocol: protocol,
	}

	sharedArgs, err := rt.paramsToArg(opts.Params)
	if err != nil {
		return nil, err
	}

	rt.baseArgs = slices.Concat(opts.Flags, fsArgs, httpArgs, sharedArgs)

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

func (rt *Binary) Run(ctx context.Context, query *source.Source, params map[string]any) (result []byte, resultErr error) {
	if err := rt.validateRunParams(params); err != nil {
		return nil, err
	}

	scriptPath, cleanup, err := resolveBinaryScript(query)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cleanupErr := cleanup(); cleanupErr != nil {
			resultErr = errors.Join(resultErr, cleanupErr)
		}
	}()

	args, err := rt.runArgs(scriptPath, params)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, rt.path, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		if len(out) != 0 {
			return nil, errors.New(string(out))
		}

		return nil, err
	}

	return out, nil
}

// ValidateSuite rejects suite execution when the selected protocol cannot
// carry the bindings between suite phases.
func (rt *Binary) ValidateSuite() error {
	if rt.protocol == ProtocolCLIDirect {
		return rt.protocol.unsupported("YAML test suites")
	}

	return nil
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

func (rt *Binary) runArgs(scriptPath string, params map[string]any) ([]string, error) {
	if err := rt.validateRunParams(params); err != nil {
		return nil, err
	}

	if rt.protocol == ProtocolCLIDirect {
		return []string{scriptPath}, nil
	}

	queryArgs, err := rt.paramsToArg(params)
	if err != nil {
		return nil, err
	}

	return slices.Concat([]string{"run", scriptPath}, rt.baseArgs, queryArgs), nil
}

func (rt *Binary) validateRunParams(params map[string]any) error {
	if rt.protocol != ProtocolCLIDirect || !hasDirectBindParameters(params) {
		return nil
	}

	return rt.protocol.unsupported("bind parameters")
}
