package sources

import "context"

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

func (a *Aggregate) Read(ctx context.Context) Stream {
	onNext := make(chan File)
	onError := make(chan error)

	srcCtx, cancel := context.WithCancel(ctx)

	go func() {
		for _, src := range a.sources {
			var done bool
			var err error

			stream := src.Read(srcCtx)

			for !done {
				select {
				case <-ctx.Done():
					return
				case e := <-stream.OnError():
					err = e
					done = true

					break
				case f, ok := <-stream.OnNext():
					if !ok {
						done = true
						break
					}

					onNext <- f
				}
			}

			if err != nil {
				cancel()

				onError <- err

				return
			}
		}
	}()

	return NewStream(onNext, onError)
}
