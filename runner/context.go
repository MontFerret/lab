package runner

import (
	"context"
	"time"
)

type Context struct {
	root   context.Context
	params Params
}

func NewContext(root context.Context, params Params) Context {
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

func (c Context) Params() Params {
	return c.params
}
