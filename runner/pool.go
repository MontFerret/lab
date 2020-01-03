package runner

type Pool struct {
	size   int
	values chan struct{}
}

func NewPool(size int) *Pool {
	return &Pool{
		size:   size,
		values: make(chan struct{}, size),
	}
}

func (pool *Pool) Capacity() int {
	return pool.size - len(pool.values)
}

func (pool *Pool) Size() int {
	return pool.size
}

func (pool *Pool) Go(f func()) {
	pool.values <- struct{}{}

	go func() {
		f()

		<-pool.values
	}()
}
