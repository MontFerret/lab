package sources

import "context"

type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

func (n Noop) Read(_ context.Context) Stream {
	f := make(chan File)
	e := make(chan error)

	defer func() {
		close(f)
		close(e)
	}()

	return Stream{
		Files: f,
		Error: e,
	}
}
