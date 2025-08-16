// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package main

import (
	"context"
	"runtime"
	"sync"
)

// Defines a work pool for executing callbacks.
//
// Work to be done is defined by a simple function, which will execute on the
// next available worker. If the queue is full, a batch of additional goroutines
// will be created up to the defined limit.
//
// Unlike a channel and infinite goroutines, this places a limit on how many can
// conversions can concurrently exist. The point is to convert audio, and memory
// nor process cycles are infinite when users are doing other work at the same
// time. Likewise, slower I/O devices such as memory cards will choke under
// heavy concurrent I/O.
//
// Unlike a channel and a fixed number of goroutines, this allows some measure
// of dynamic scaling. A low limit can restrict resource usage during export. A
// high limit can ramp up more concurrent exports if you're willing to dedicated
// excessive resources, but don't always export such a large collection.
type WorkPool struct {
	ctx    context.Context    // Used for shutdown of the pool.
	cancel context.CancelFunc // Used for shutdown of the pool.
	buffer int                // Buffer size for queue.
	size   int                // Number of goroutines in the pool.
	limit  int                // Max value for size.
	wg     sync.WaitGroup     // Used for shutdown of the pool.
	mutex  sync.Mutex         // Protects the size field.
	queue  chan func()        // Channel of tasks for the goroutines.
}

// Creates a new work pool. Call Start() to spawn the initial workers and use
// Add() to push callbacks onto the work queue for execution.
//
// A maximum number of goroutines can be specified with [limit]. If 0, the limit
// is [runtime.NumCPU].
//
// The queue size will be set to [buffer], or a default value if 0 was provided.
func NewWorkPool(parent context.Context, limit int, buffer int) *WorkPool {
	ctx, cancel := context.WithCancel(parent)
	if limit == 0 {
		limit = runtime.NumCPU()
	}
	if buffer == 0 {
		buffer = max(limit, 100)
	}
	return &WorkPool{
		ctx:    ctx,
		cancel: cancel,
		limit:  limit,
		buffer: buffer,
		queue:  make(chan func(), buffer),
	}
}

// Spawns a set of workers. Up to [runtime.NumCPU] or the pool limit will be
// created. Additional goroutines will be generated as necessary up to the
// limit.
func (p *WorkPool) Start() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.init()
}

// Peforms initialization of the pool. This must be called while holding
// p.mutex.
func (p *WorkPool) init() {
	if p.size > 0 {
		panic("init called on running WorkPool!")
	}
	p.queue = make(chan func(), p.buffer)
	ncpu := runtime.NumCPU()
	for i := 0; i < ncpu && i < p.limit; i++ {
		p.wg.Add(1)
		p.size++
		go p.worker()
	}
}

// Stop all workers and abort tasks remaining in the queue. This will block
// until everyone finishes their current item and halts. If the parent context
// is closed, this occurs automatically. After calling this returns, you must
// call [Start] before it is possible to add any more items to the queue.
func (p *WorkPool) Stop() {
	// Acquire the mutex to prevent anyone calling Add().
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.size == 0 {
		panic("WorkPool.Stop called when already stopped")
	}

	// Halt the workers at their next tick.
	p.cancel()
	p.wg.Wait()
	p.size = 0

	// There may be items remaining in the queue. To ensure they're subject to
	// GC, they or the queue must go. Since nil'ing the queue would be a data
	// race with concurrent Adds (not that you should be doing that, if you
	// called Stop :), we close and drain the queue.
	close(p.queue)
	for range p.queue {
		// We don't want to execute, just ensure the channel doesn't retain the
		// memory.
	}
}

// Drain the queue and halt all workers. This can be used to wait for the
// completion of currently queued callbacks.
func (p *WorkPool) Wait() {
	// Workers will halt once the queue drains.
	p.mutex.Lock()
	defer p.mutex.Unlock()
	close(p.queue)
	p.wg.Wait()
	p.size = 0
	// Restart the queue and initial goroutines. We perform this with a separate
	// init method, because if we unlocked the mutex in order to call Start():
	// if Add()->expand() was called asyncronously with Wait(), there would be a
	// data race where expand could see the queue is stopped (p.size==0) and
	// when the it exists, depending on which goroutine obtained the lock first.
	p.init()
}

// Add a callback to the work queue. If the queue is full, additional goroutines
// will be spawned up to the limit. By default, the queue is
func (p *WorkPool) Add(fn func()) {
	p.expand()
	p.queue <- fn
}

// Possibly expands the work pool. Up to 4 workers are created if the queue is
// full, provided the limit has not been reached.
func (p *WorkPool) expand() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.size == 0 {
		panic("WorkPool is not running")
	}
	if p.size == p.limit {
		// Pool can't grow any further.
		return
	}

	if p.Remaining() > 0 {
		return
	}

	// Up to the limit, or this many new go routines.
	growth := min(4, p.limit-p.size)

	for range growth {
		p.wg.Add(1)
		p.size++
		go p.worker()
	}
}

// Returns the approximate amount of queue space remaining.
func (p *WorkPool) Remaining() int {
	return cap(p.queue) - len(p.queue)
}

// Returns a percentage of how full the queue is.
//
// Like Remaining, but expressed as a percentage of usage rather than the count
// of available slots. E.g., "6.0" means the queue is 6% full but Remaining()
// might be saying 94 slots out of a 100 are free.
func (p *WorkPool) PercentFull() float64 {
	return float64(len(p.queue)) / float64(cap(p.queue)) * 100.0
}

// Return the current size of the work pool.
func (p *WorkPool) Size() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.size
}

// Returns the maximum number of workers allowed.
func (p *WorkPool) Limit() int {
	return p.limit
}

func (p *WorkPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			// Pool is shutting down.
			return
		case fn, ok := <-p.queue:
			if !ok {
				// The queue is closed.
				return
			}
			fn()
		}
	}
}
