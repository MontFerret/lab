package runtime

import (
	"context"
	"github.com/pkg/errors"
)

type Remote struct {
	url string
}

func NewRemote(url string) *Remote {
	return &Remote{url: url}
}

func (r Remote) Run(ctx context.Context, program string, params map[string]interface{}) ([]byte, error) {
	return nil, errors.New("not implemented")
}
