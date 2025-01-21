package wopo

// PoolOption option for NewPool
type PoolOption[T, V any] func(p *Pool[T, V])

// WithWorkerCount sets the number of workers. Panics if n <= 0
func WithWorkerCount[T, V any](n int) PoolOption[T, V] {
	if n <= 0 {
		panic("number of workers must be positive")
	}

	return func(wp *Pool[T, V]) {
		wp.workerCount = n
	}
}

// WithTaskBufferSize sets the buffer size of the input channel
func WithTaskBufferSize[T, V any](s int) PoolOption[T, V] {
	return func(wp *Pool[T, V]) {
		wp.inCh = make(chan task[T], s)
	}
}

// WithResultBufferSize sets the buffer size of the output channel.
// The value s = -1 means that the pool does not need to return the result.
// In this case, the results channel is initialized closed, and the handler result is ignored
func WithResultBufferSize[T, V any](s int) PoolOption[T, V] {
	return func(wp *Pool[T, V]) {
		if s == -1 {
			wp.noResults = true
			close(wp.outCh)
			return
		}
		wp.outCh = make(chan Result[V], s)
	}
}
