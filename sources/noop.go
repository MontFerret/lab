package sources

import (
	"context"
	"net/url"
)

type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

func (n Noop) Read(_ context.Context) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	defer func() {
		close(onNext)
		close(onError)
	}()

	return onNext, onError
}

func (n Noop) Resolve(ctx context.Context, _ url.URL) (<-chan File, <-chan Error) {
	return n.Read(ctx)
}
