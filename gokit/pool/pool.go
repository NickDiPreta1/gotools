// Package pool provides a worker pool implementation for concurrent job processing.
package pool

import (
	"context"
	"sync"
)

// Job represents a unit of work to be processed by the worker pool.
// Each job has a unique ID, content to process, and a function to execute.
type Job struct {
	ID      int
	Content []byte
	Func    func([]byte) ([]byte, error)
}

// Result represents the outcome of processing a job.
// It contains the job ID, processed content, and any error that occurred.
type Result struct {
	JobID   int
	Content []byte
	Error   error
}

// Pool manages a pool of workers that process jobs concurrently.
// It maintains channels for job submission and result collection,
// and uses a WaitGroup to track worker lifecycle.
type Pool struct {
	workerCount int
	jobs        chan Job
	results     chan Result
	wg          sync.WaitGroup
}

// New creates a new worker pool.
// workerCount specifies the number of worker goroutines to spawn.
// bufferSize sets the capacity of both the jobs and results channels.
// Returns a pointer to the newly created Pool.
func New(workerCount int, bufferSize int) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobs:        make(chan Job, bufferSize),
		results:     make(chan Result, bufferSize),
	}
}

// worker is the goroutine that processes jobs from the jobs channel.
// It runs in a loop, selecting between receiving jobs and context cancellation.
// When a job is received, it executes the job's function and sends the result.
// The worker terminates when the jobs channel is closed or the context is cancelled.
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			result, err := job.Func(job.Content)
			if err != nil {
				p.results <- Result{
					JobID: job.ID,
					Error: err,
				}
				continue
			}
			p.results <- Result{
				JobID:   job.ID,
				Content: result,
			}

		case <-ctx.Done():
			return
		}
	}
}

// Start initializes and starts all worker goroutines.
// It spawns workerCount number of workers that will process jobs concurrently.
// Returns a read-only channel that will emit results as jobs are completed.
// The caller should consume from this channel to receive job results.
func (p *Pool) Start(ctx context.Context) <-chan Result {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
	return p.results
}

// Submit adds a job to the pool for processing.
// The job will be picked up by an available worker.
// This call will block if the jobs channel buffer is full.
func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

// Shutdown gracefully shuts down the worker pool.
// It closes the jobs channel to signal workers to stop accepting new jobs,
// waits for all workers to complete their current jobs,
// and then closes the results channel.
// After calling Shutdown, no new jobs should be submitted.
func (p *Pool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
}
