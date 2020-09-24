package sources

import "context"

type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

func (n Noop) Read(_ context.Context) Stream {
	onNext := make(chan File)
	onError := make(chan error)

	defer func() {
		close(onNext)
		close(onError)
	}()

	return NewStream(onNext, onError)
}
