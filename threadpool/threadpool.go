package threadpool

import (
	"runtime/debug"

	"github.com/cockroachdb/errors"
	"github.com/gammazero/workerpool"
)

// Pool specifies how to use
// this internal threadpool implementation.
type Pool interface {
	Run(task func())
	QueueSize() int64
	Stop()
}

// Impl is an production grade threadpool implementation.
type Impl struct {
	wp *workerpool.WorkerPool
}

// NewThreadPool returns a pool implementation.
func NewThreadPool(size int) (Pool, error) {
	pool := workerpool.New(size)

	if pool == nil {
		return nil, errors.WithStack(errors.New("could not create worker pool"))
	}

	return &Impl{
		wp: pool,
	}, nil
}

// Run schedules to given task to
// next available worker.
func (e *Impl) Run(task func()) {
	e.wp.Submit(func() {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
			}
		}()

		task()
	})
}

// Stop waits for the ingoing tasks to finished
// and closes the threadpool.
func (e *Impl) Stop() {
	e.wp.StopWait()
}

// QueueSize returns the number of
// waiting message in the underlying queue.
func (e *Impl) QueueSize() int64 {
	return int64(e.wp.WaitingQueueSize())
}
