// Package batcher implements the batching resiliency pattern for Go.
package batcher

import (
	"sync"
	"time"
)

type work struct {
	param  interface{}
	future chan error
}

// Batcher implements the batching resiliency pattern
type Batcher struct {
	timeout   time.Duration
	prefilter func(interface{}) error

	lock         sync.Mutex
	submit       chan *work
	doWork       func([]interface{}) error
	batchCounter sync.WaitGroup
	flushTimer   *time.Timer
}

// New constructs a new batcher that will batch all calls to Run that occur within
// `timeout` time before calling doWork just once for the entire batch. The doWork
// function must be safe to run concurrently with itself as this may occur, especially
// when the doWork function is slow, or the timeout is small.
func New(timeout time.Duration, doWork func([]interface{}) error) *Batcher {
	return &Batcher{
		timeout: timeout,
		doWork:  doWork,
	}
}

// Run runs the work function with the given parameter, possibly
// including it in a batch with other calls to Run that occur within the
// specified timeout. It is safe to call Run concurrently on the same batcher.
func (b *Batcher) Run(param interface{}) error {
	if b.prefilter != nil {
		if err := b.prefilter(param); err != nil {
			return err
		}
	}

	if b.timeout == 0 {
		return b.doWork([]interface{}{param})
	}

	w := &work{
		param:  param,
		future: make(chan error, 1),
	}

	b.submitWork(w)

	return <-w.future
}

// Prefilter specifies an optional function that can be used to run initial checks on parameters
// passed to Run before being added to the batch. If the prefilter returns a non-nil error,
// that error is returned immediately from Run and the batcher is not invoked. A prefilter
// cannot safely be specified for a batcher if Run has already been invoked. The filter function
// specified must be concurrency-safe.
func (b *Batcher) Prefilter(filter func(interface{}) error) {
	b.prefilter = filter
}

func (b *Batcher) submitWork(w *work) {
	b.lock.Lock()
	defer b.lock.Unlock()

	// kick off a new batch if needed
	if b.submit == nil {
		b.batchCounter.Add(1)
		b.submit = make(chan *work, 4)
		go b.batch(b.submit)
		b.flushTimer = time.AfterFunc(b.timeout, b.flushCurrentBatch)
	}

	// then add this work to the current batch
	b.submit <- w
}

func (b *Batcher) batch(input <-chan *work) {
	defer b.batchCounter.Done()

	var params []interface{}
	var futures []chan error

	for work := range input {
		params = append(params, work.param)
		futures = append(futures, work.future)
	}

	ret := b.doWork(params)

	for _, future := range futures {
		future <- ret
		close(future)
	}
}

// Shutdown flushes and executes any pending batches. If wait is true, it also waits for the pending batches
// to finish executing before it returns. This can be used to avoid waiting for the timeout to expire when
// gracefully shutting down your application. Calling Run at any point after calling Shutdown will lead to
// undefined behaviour.
func (b *Batcher) Shutdown(wait bool) {
	b.flushCurrentBatch()

	if wait {
		b.batchCounter.Wait()
	}
}

func (b *Batcher) flushCurrentBatch() {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.submit == nil {
		return
	}

	// stop the timer to avoid spurious flushes and trigger immediate cleanup in case this flush was
	// triggered manually by a call to Shutdown (it has to happen inside the lock, so it can't be done
	// in the Shutdown method directly)
	b.flushTimer.Stop()

	close(b.submit)
	b.submit = nil
}
