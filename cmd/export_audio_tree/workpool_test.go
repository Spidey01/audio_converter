package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

type task struct {
	ctx    context.Context
	cancel context.CancelFunc
	tid    int
}

func newTask(parent context.Context, tid int) *task {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	return &task{
		tid:    tid,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (task *task) String() string {
	return fmt.Sprintln("task", task.tid)
}

func (task *task) run(t *testing.T) {
	t.Logf("+++ %s started", task)
	defer t.Logf("--- %s stopped", task)
	defer task.cancel()
	for {
		select {
		case <-t.Context().Done():
			t.Logf("=== Test exiting")
			return
		case <-task.ctx.Done():
			t.Logf("=== Task %d complete", task.tid)
			return
		}
	}
}

func TestWorkPool(t *testing.T) {
	t.Run("Basic attributes", func(t *testing.T) {
		limit := 1024
		buffer := 10
		t.Logf("NewWorkPool(_, %d, %d)", limit, buffer)
		p := NewWorkPool(t.Context(), limit, buffer)
		if n := p.Limit(); n != limit {
			t.Errorf("Bad limit: actual: %d expected: %d", n, limit)
		}
		if n := p.Remaining(); n != buffer {
			t.Errorf("Bad buffer remaining: actual: %d expected: %d", n, buffer)
		}
		if f := p.PercentFull(); f < -0.001 || f > +0.001 {
			t.Errorf("Bad percent full: actual: %f expected 0.0", f)
		}
		if s := p.Size(); s != 0 {
			t.Errorf("Bad size: actual: %d expected: 0", s)
		}
	})
	t.Run("Start stop", func(t *testing.T) {
		t.Logf("Creating pool")
		pool := NewWorkPool(t.Context(), 0, 0)
		t.Logf("Starting pool")
		pool.Start()

		size := min(runtime.NumCPU()*2, pool.Remaining())
		var tasks []*task
		t.Logf("Adding some tasks")
		for i := range size {
			task := newTask(t.Context(), i)
			tasks = append(tasks, task)
			// Because items in the queue may or may not run, we can't include
			// them in the wait group -- if they don't run, obviously they won't
			// call done.
			pool.Add(func() {
				task.run(t)
			})
		}

		var wg sync.WaitGroup

		// Put the stop in its own goroutine.
		wg.Add(1)
		go func() {
			t.Logf("Stopping pool")
			pool.Stop()
			t.Logf("All tasks completed")
			wg.Done()
		}()

		// Mark each task complete after a random interval in microseconds.
		for _, task := range tasks {
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(100)))
			t.Logf("Marking %s complete", task)
			task.cancel()
		}
		t.Logf("Waiting for all goroutines to finish")
		wg.Wait()

		// Mostly for completeness sake, let's make sure we can restart the pool
		t.Logf("Restarting the pool")
		pool.Start()
		wg.Add(1)
		pool.Add(func() {
			t.Log("Restarting the pool worked")
			wg.Done()
		})
		pool.Stop()
		wg.Wait()
	})

	t.Run("start wait", func(t *testing.T) {
		t.Logf("Create pool")
		pool := NewWorkPool(t.Context(), 0, 0)
		t.Logf("Start pool")
		pool.Start()
		defer pool.Stop()

		ctx, done := context.WithCancel(t.Context())
		task := newTask(ctx, 1)
		pool.Add(func() { task.run(t) })

		interval := time.Duration(rand.Intn(500)) * time.Microsecond
		t.Logf("Marking %s done in %v", task, interval)
		go func() {
			time.Sleep(interval)
			done()
		}()

		t.Logf("Waiting for %s to finish", task)
		pool.Wait()

		t.Logf("Restarting pool")
		var wg sync.WaitGroup
		wg.Add(1)
		pool.Add(func() {
			t.Log("Verifying add after wait doesn't crash")
			wg.Done()
		})
		wg.Wait()
		t.Logf("Adding task after wait worked")
	})

}
