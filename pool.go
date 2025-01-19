package wopo

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

var (
	// ErrRecoveredFromPanic error indicating that a panic occurred during the task execution
	ErrRecoveredFromPanic = errors.New("recovered from panic")
)

type Handler[T, V any] func(ctx context.Context, in T) (out V, err error)

// task data and context for the task passed to the Handler
type task[T any] struct {
	Ctx  context.Context
	Data T
}

// Result result data and error after Handler execution.
// Panics that occur in the Handler turn into an error ErrRecoveredFromPanic
type Result[V any] struct {
	Err  error
	Data V
}

// Pool structure of the worker pool
type Pool[T, V any] struct {
	workerCount int
	handler     Handler[T, V]

	inCh  chan task[T]
	outCh chan Result[V]

	stopWg sync.WaitGroup
	stopCh chan struct{}
}

// NewPool creates a pool structure for workers with a handler,
// but does not launch the workers themselves. Parameterized via PoolOption:
//
// WithWorkerCount sets the number of workers. Panics if n <= 0. Default is 3
//
// WithTaskBufferSize sets the buffer size of the input channel. Default is 0
//
// WithResultBufferSize sets the buffer size of the output channel. Default is 0
func NewPool[T, V any](handler Handler[T, V], opts ...PoolOption[T, V]) *Pool[T, V] {
	p := &Pool[T, V]{
		handler:     handler,
		workerCount: 3,
		stopWg:      sync.WaitGroup{},
		inCh:        make(chan task[T]),
		outCh:       make(chan Result[V]),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Start launches pool handlers
func (p *Pool[T, V]) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.stopWg.Add(1)
		go p.workLoop()
	}
}

// Stop closes the input channel and is blocked while waiting for the completion of running handlers.
// After calling Stop, you MUST NOT use the PushTask method to run a task
func (p *Pool[T, V]) Stop() {
	close(p.inCh)
	p.stopWg.Wait()
	close(p.outCh)
}

// PushTask puts the task in the local queue for processing by the pool.
// The call is blocked if the input buffer is completely full.
// You MUST NOT use this method after calling the Stop method.
func (p *Pool[T, V]) PushTask(ctx context.Context, taskData T) {
	p.inCh <- task[T]{Ctx: ctx, Data: taskData}
}

func (p *Pool[T, V]) workLoop() {
	defer p.stopWg.Done()

	for t := range p.inCh {
		p.outCh <- p.handle(t)
	}
}

// Results the channel in which the result of the task processing is recorded.
// Closes after the pool is completely stopped
func (p *Pool[T, V]) Results() <-chan Result[V] {
	return p.outCh
}

func (p *Pool[T, V]) handle(task task[T]) (res Result[V]) {
	defer func() {
		if r := recover(); r != nil {
			res = Result[V]{Err: fmt.Errorf("%w: %s", ErrRecoveredFromPanic, string(debug.Stack()))}
			return
		}
	}()

	resData, err := p.handler(task.Ctx, task.Data)
	return Result[V]{Data: resData, Err: err}
}
