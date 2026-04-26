package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/MontFerret/ferret/v2/pkg/source"
)

type Binary struct {
	path         string
	sharedParams map[string]any
	rawFlags     []string
}

func NewBinary(path string, params map[string]any) (*Binary, error) {
	sharedParams, rawFlags, err := splitBinaryRuntimeParams(params)

	if err != nil {
		return nil, err
	}

	return &Binary{path, sharedParams, rawFlags}, nil
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
	args := make([]string, 0, 10)
	args = append(args, rt.rawFlags...)

	sharedArgs, err := rt.paramsToArg(rt.sharedParams)

	if err != nil {
		return nil, err
	}

	queryArgs, err := rt.paramsToArg(params)

	if err != nil {
		return nil, err
	}

	args = append(args, sharedArgs...)
	args = append(args, queryArgs...)

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

func (rt *Binary) paramsToArg(params map[string]any) ([]string, error) {
	args := make([]string, 0, len(params))

	for k, v := range params {
		j, err := json.Marshal(v)

		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to serialize parameter: %s", k))
		}

		args = append(args, fmt.Sprintf("--param=%s:%s", k, j))
	}

	return args, nil
}

func splitBinaryRuntimeParams(params map[string]any) (map[string]any, []string, error) {
	if len(params) == 0 {
		return params, nil, nil
	}

	sharedParams := make(map[string]any, len(params))
	var rawFlags []string

	for k, v := range params {
		if k != "flags" {
			sharedParams[k] = v

			continue
		}

		flags, err := toStringSlice(v)

		if err != nil {
			return nil, nil, errors.Wrap(err, "invalid type of flags (expected array of strings)")
		}

		rawFlags = flags
	}

	return sharedParams, rawFlags, nil
}

func toStringSlice(value any) ([]string, error) {
	switch values := value.(type) {
	case []string:
		return append([]string(nil), values...), nil
	case []any:
		res := make([]string, 0, len(values))

		for _, v := range values {
			str, ok := v.(string)

			if !ok {
				return nil, errors.New("expected string value")
			}

			res = append(res, str)
		}

		return res, nil
	default:
		return nil, errors.New("expected array")
	}
}
