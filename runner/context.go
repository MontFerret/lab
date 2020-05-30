package runner

import (
	"context"
	"time"

	"github.com/MontFerret/lab/testing"
)

type Context struct {
	root   context.Context
	params testing.Params
}

func NewContext(root context.Context, params testing.Params) Context {
	return Context{
		root:   root,
		params: params,
	}
}

func (c Context) Deadline() (time.Time, bool) {
	return c.root.Deadline()
}

func (c Context) Done() <-chan struct{} {
	return c.root.Done()
}

func (c Context) Err() error {
	return c.root.Err()
}

func (c Context) Value(key interface{}) interface{} {
	return c.root.Value(key)
}

func (c Context) Params() testing.Params {
	return c.params
}
