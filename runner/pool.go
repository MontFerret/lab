package runner

type Pool struct {
	size   uint64
	values chan struct{}
}

func NewPool(size uint64) *Pool {
	return &Pool{
		size:   size,
		values: make(chan struct{}, size),
	}
}

func (pool *Pool) Capacity() uint64 {
	return pool.size - uint64(len(pool.values))
}

func (pool *Pool) Size() uint64 {
	return pool.size
}

func (pool *Pool) Go(f func()) {
	pool.values <- struct{}{}

	go func() {
		f()

		<-pool.values
	}()
}
