package workerpool

import (
	"log"
	"sync"
)

// Task represents a unit of work to be processed by the pool.
type Task func()

// Pool is a bounded worker pool that processes tasks from a buffered channel.
type Pool struct {
	tasks chan Task
	wg    sync.WaitGroup
}

// New creates a worker pool with the given number of workers and queue size.
// Workers start immediately and process tasks until Stop is called.
func New(workers, queueSize int) *Pool {
	p := &Pool{
		tasks: make(chan Task, queueSize),
	}
	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer p.wg.Done()
			for task := range p.tasks {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("workerpool: task panicked: %v", r)
						}
					}()
					task()
				}()
			}
		}()
	}
	return p
}

// Submit adds a task to the pool. If the queue is full, the task is dropped
// and a warning is logged (non-blocking to avoid slowing down request handlers).
func (p *Pool) Submit(task Task) {
	select {
	case p.tasks <- task:
	default:
		log.Printf("workerpool: queue full, task dropped")
	}
}

// Stop closes the task channel and waits for all workers to finish.
func (p *Pool) Stop() {
	close(p.tasks)
	p.wg.Wait()
}
