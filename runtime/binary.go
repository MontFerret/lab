package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

type Binary struct {
	path string
}

func NewBinary(path string, params map[string]interface{}) (*Binary, error) {
	return &Binary{path: path}, nil
}

func (b *Binary) Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
	p, err := b.paramsToArg(params)

	if err != nil {
		return nil, err
	}

	q := &bytes.Buffer{}
	q.WriteString(query)

	cmd := exec.CommandContext(ctx, b.path, p)
	cmd.Stdin = q

	out, err := cmd.CombinedOutput()

	if err != nil {
		if len(out) != 0 {
			return nil, errors.New(string(out))
		}

		return nil, err
	}

	return out, nil
}

func (b *Binary) paramsToArg(params map[string]interface{}) (string, error) {
	var buff bytes.Buffer

	for k, v := range params {
		j, err := json.Marshal(v)

		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to serialize parameter: %s", k))
		}

		buff.WriteString("--param=")
		buff.WriteString(k)
		buff.WriteString(":")
		buff.Write(j)
		buff.WriteString(" ")
	}

	return buff.String(), nil
}
