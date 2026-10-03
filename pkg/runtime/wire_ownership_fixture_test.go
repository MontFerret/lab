package runtime

type closeCallback struct {
	fn func() error
}

func (c *closeCallback) Close() error {
	return c.fn()
}
