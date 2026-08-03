package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
)

var (
	ErrPoolIsDone        = errors.New("worker pool is done")
	ErrNilTask           = errors.New("nil task")
	ErrPoolIsStop        = errors.New("worker pool is stopped")
	ErrPoolAlreadyClosed = errors.New("pool is already closed")
)

// Task represents a unit of work to be executed by the worker pool.
type Task struct {
	ctx context.Context
	fn  func(context.Context) error
}

// WorkerPool manages a fixed number of goroutines that process tasks
// from a buffered channel. It supports graceful shutdown and error collection.
type WorkerPool struct {
	tasks   chan Task
	errs    chan error
	dropped atomic.Int64

	done chan struct{}
	stop chan struct{}

	once sync.Once
	wg   sync.WaitGroup
}

// New creates and starts a new WorkerPool with the specified configuration.
// Workers are spawned immediately and begin listening for tasks.
func New(cfg *config.Config) *WorkerPool {
	wp := &WorkerPool{
		tasks: make(chan Task, cfg.WorkerPool.TaskBuf),
		errs:  make(chan error, cfg.WorkerPool.ErrBuf),
		done:  make(chan struct{}),
		stop:  make(chan struct{}),
	}

	wp.wg.Add(int(cfg.WorkerPool.Workers))
	for range cfg.WorkerPool.Workers {
		go wp.worker()
	}

	return wp
}

// Submit enqueues a task for asynchronous execution.
//
// It returns immediately if the task is successfully queued. If the pool is
// stopped, the context is canceled, or the function is nil, an appropriate
// error is returned without blocking.
func (wp *WorkerPool) Submit(ctx context.Context, fn func(context.Context) error) error {
	if fn == nil {
		return ErrNilTask
	}

	select {
	case <-wp.done:
		return ErrPoolIsDone
	case <-wp.stop:
		return ErrPoolIsStop
	default:
	}

	task := Task{
		ctx: ctx,
		fn:  fn,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-wp.stop:
		return ErrPoolIsStop
	case wp.tasks <- task:
		return nil
	}
}

// Shutdown initiates a graceful shutdown of the worker pool.
//
// It stops accepting new tasks, waits for all in-flight tasks to complete,
// and respects the provided context deadline. If the context expires before
// all workers finish, Shutdown returns the context error but continues
// waiting in the background to prevent goroutine leaks.
//
// Returns ErrPoolAlreadyClosed if called more than once.
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	closed := false
	wp.once.Do(func() {
		closed = true
		close(wp.stop)
	})

	if !closed {
		return ErrPoolAlreadyClosed
	}

	waited := make(chan struct{})
	go func() {
		defer close(waited)
		wp.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		close(wp.done)
		wp.wg.Wait()
		return ctx.Err()
	case <-waited:
		return nil
	}
}

// Errors returns a read-only channel for receiving task execution errors.
//
// Consumers should read from this channel continuously to prevent blocking
// workers when the error buffer is full.
func (wp *WorkerPool) Errors() <-chan error {
	return wp.errs
}

// Dropped returns the total number of errors that could not be sent
// to the error channel due to buffer overflow.
func (wp *WorkerPool) Dropped() int64 {
	return wp.dropped.Load()
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.done:
			return
		case <-wp.stop:
			wp.drainEndExit()
			return
		case t := <-wp.tasks:
			wp.runTask(t)
		}
	}
}

func (wp *WorkerPool) drainEndExit() {
	for {
		select {
		case <-wp.done:
			return
		case t := <-wp.tasks:
			wp.runTask(t)
		default:
			return
		}
	}
}

func (wp *WorkerPool) runTask(t Task) {
	defer func() {
		if err := recover(); err != nil {
			wp.sendErr(fmt.Errorf("panic recovered: %v", err))
		}
	}()

	if err := t.ctx.Err(); err != nil {
		wp.sendErr(fmt.Errorf("ctx of task is done: %w", err))
	}

	if err := t.fn(t.ctx); err != nil {
		wp.sendErr(fmt.Errorf("can't exec func: %w", err))
	}
}

func (wp *WorkerPool) sendErr(err error) {
	select {
	case wp.errs <- err:
		return
	default:
	}

	wp.dropped.Add(1)
}
