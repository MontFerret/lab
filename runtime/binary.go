package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Binary struct {
	path         string
	cdpAddress   string
	sharedParams map[string]interface{}
}

func NewBinary(path string, cdpAddress string, params map[string]interface{}) (*Binary, error) {
	return &Binary{path, cdpAddress, params}, nil
}

func (rt *Binary) Version(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, rt.path, "--version")

	out, err := cmd.CombinedOutput()

	if err != nil {
		if len(out) != 0 {
			return "", errors.New(string(out))
		}

		return "", err
	}

	return strings.ReplaceAll(string(out), "\n", ""), nil
}

func (rt *Binary) Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
	args := make([]string, 0, 10)
	args = append(args, "--cdp="+rt.cdpAddress)

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
	q.WriteString(query)

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

func (rt *Binary) paramsToArg(params map[string]interface{}) ([]string, error) {
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
