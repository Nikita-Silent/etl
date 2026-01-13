package workers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Pool represents a worker pool that limits concurrent goroutines
type Pool struct {
	// semaphore channel to limit concurrent workers
	semaphore chan struct{}
	// wait group to track active workers
	wg sync.WaitGroup
	// worker count for logging
	maxWorkers int
}

// NewPool creates a new worker pool with the specified number of workers
func NewPool(maxWorkers int) *Pool {
	if maxWorkers <= 0 {
		maxWorkers = 10 // Default to 10 workers
	}

	slog.Info("Creating worker pool",
		"max_workers", maxWorkers,
		"event", "worker_pool_init",
	)

	return &Pool{
		semaphore:  make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

// Submit submits a task to the worker pool
// Blocks if all workers are busy until a worker becomes available
func (p *Pool) Submit(ctx context.Context, task func() error) error {
	// Check if context is canceled before acquiring semaphore
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Acquire semaphore (blocks if pool is full)
	select {
	case p.semaphore <- struct{}{}:
		// Successfully acquired semaphore
	case <-ctx.Done():
		return ctx.Err()
	}

	// Increment wait group
	p.wg.Add(1)

	// Execute task in goroutine
	go func() {
		defer func() {
			// Release semaphore
			<-p.semaphore
			// Decrement wait group
			p.wg.Done()
		}()

		// Execute the task
		if err := task(); err != nil {
			slog.Debug("Worker task completed with error",
				"error", err.Error(),
				"event", "worker_task_error",
			)
		}
	}()

	return nil
}

// Wait waits for all workers to complete
func (p *Pool) Wait() {
	p.wg.Wait()
	slog.Debug("All workers completed",
		"event", "worker_pool_complete",
	)
}

// Stats returns current pool statistics
type Stats struct {
	MaxWorkers    int
	ActiveWorkers int
	QueuedTasks   int
}

// GetStats returns current pool statistics
func (p *Pool) GetStats() Stats {
	activeWorkers := len(p.semaphore)
	return Stats{
		MaxWorkers:    p.maxWorkers,
		ActiveWorkers: activeWorkers,
		QueuedTasks:   0, // Not tracked in this simple implementation
	}
}

// SubmitMany submits multiple tasks to the worker pool
// Returns the number of successfully submitted tasks and any error
func (p *Pool) SubmitMany(ctx context.Context, tasks []func() error) (int, error) {
	submitted := 0
	for i, task := range tasks {
		if err := p.Submit(ctx, task); err != nil {
			return submitted, fmt.Errorf("failed to submit task %d: %w", i, err)
		}
		submitted++
	}
	return submitted, nil
}

// SubmitFunc is a convenience method that wraps a function with no return value
func (p *Pool) SubmitFunc(ctx context.Context, fn func()) error {
	return p.Submit(ctx, func() error {
		fn()
		return nil
	})
}
