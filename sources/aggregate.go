package sources

import (
	"context"
	"net/url"
)

type Aggregate struct {
	sources []Source
}

func NewAggregate(sources ...Source) *Aggregate {
	return &Aggregate{sources}
}

func (a *Aggregate) Add(src Source) {
	if src == nil {
		return
	}

	a.sources = append(a.sources, src)
}

func (a *Aggregate) Read(ctx context.Context) (<-chan File, <-chan Error) {
	onNext := make(chan File)
	onError := make(chan Error)

	go func() {
		for _, src := range a.sources {
			var done bool

			next, err := src.Read(ctx)

			for !done {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-err:
					if !ok {
						done = true
						break
					}

					onError <- e
				case f, ok := <-next:
					if !ok {
						done = true
						break
					}

					onNext <- f
				}
			}
		}

		close(onNext)
		close(onError)
	}()

	return onNext, onError
}

func (a *Aggregate) Resolve(_ context.Context, _ *url.URL) (<-chan File, <-chan Error) {
	return nil, nil
}
