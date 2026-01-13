package workers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	tests := []struct {
		name            string
		maxWorkers      int
		expectedWorkers int
	}{
		{"valid size", 5, 5},
		{"zero size uses default", 0, 10},
		{"negative size uses default", -1, 10},
		{"large size", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(tt.maxWorkers)
			if pool == nil {
				t.Fatal("NewPool returned nil")
			}
			if pool.maxWorkers != tt.expectedWorkers {
				t.Errorf("maxWorkers = %d, want %d", pool.maxWorkers, tt.expectedWorkers)
			}
			if cap(pool.semaphore) != tt.expectedWorkers {
				t.Errorf("semaphore capacity = %d, want %d", cap(pool.semaphore), tt.expectedWorkers)
			}
		})
	}
}

func TestPoolSubmit(t *testing.T) {
	pool := NewPool(2)
	ctx := context.Background()

	executed := false
	task := func() error {
		executed = true
		return nil
	}

	err := pool.Submit(ctx, task)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	pool.Wait()

	if !executed {
		t.Error("task was not executed")
	}
}

func TestPoolSubmitWithError(t *testing.T) {
	pool := NewPool(2)
	ctx := context.Background()

	expectedErr := errors.New("task error")
	task := func() error {
		return expectedErr
	}

	err := pool.Submit(ctx, task)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	pool.Wait()
	// No error should propagate from Submit itself
}

func TestPoolConcurrentSubmit(t *testing.T) {
	maxWorkers := 5
	pool := NewPool(maxWorkers)
	ctx := context.Background()

	taskCount := 50
	var counter int64

	for i := 0; i < taskCount; i++ {
		err := pool.Submit(ctx, func() error {
			atomic.AddInt64(&counter, 1)
			time.Sleep(time.Millisecond * 10)
			return nil
		})
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	pool.Wait()

	if counter != int64(taskCount) {
		t.Errorf("counter = %d, want %d", counter, taskCount)
	}
}

func TestPoolMaxConcurrency(t *testing.T) {
	maxWorkers := 3
	pool := NewPool(maxWorkers)
	ctx := context.Background()

	var activeTasks int64
	var maxConcurrent int64
	var mu sync.Mutex

	taskCount := 10

	for i := 0; i < taskCount; i++ {
		err := pool.Submit(ctx, func() error {
			current := atomic.AddInt64(&activeTasks, 1)

			mu.Lock()
			if current > maxConcurrent {
				maxConcurrent = current
			}
			mu.Unlock()

			time.Sleep(time.Millisecond * 50)
			atomic.AddInt64(&activeTasks, -1)
			return nil
		})
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	pool.Wait()

	if maxConcurrent > int64(maxWorkers) {
		t.Errorf("max concurrent tasks = %d, want <= %d", maxConcurrent, maxWorkers)
	}
	if maxConcurrent < 1 {
		t.Error("no tasks were executed concurrently")
	}
}

func TestPoolContextCancellation(t *testing.T) {
	pool := NewPool(1)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	task := func() error {
		t.Error("task should not be executed")
		return nil
	}

	err := pool.Submit(ctx, task)
	if err == nil {
		t.Error("Submit should fail with canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestPoolWait(t *testing.T) {
	pool := NewPool(2)
	ctx := context.Background()

	taskCount := 10
	completed := make([]bool, taskCount)
	var mu sync.Mutex

	for i := 0; i < taskCount; i++ {
		idx := i
		err := pool.Submit(ctx, func() error {
			time.Sleep(time.Millisecond * 10)
			mu.Lock()
			completed[idx] = true
			mu.Unlock()
			return nil
		})
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	pool.Wait()

	// All tasks should be completed
	for i, done := range completed {
		if !done {
			t.Errorf("task %d was not completed", i)
		}
	}
}

func TestPoolGetStats(t *testing.T) {
	maxWorkers := 5
	pool := NewPool(maxWorkers)

	stats := pool.GetStats()
	if stats.MaxWorkers != maxWorkers {
		t.Errorf("stats.MaxWorkers = %d, want %d", stats.MaxWorkers, maxWorkers)
	}
	if stats.ActiveWorkers != 0 {
		t.Errorf("stats.ActiveWorkers = %d, want 0", stats.ActiveWorkers)
	}
}

func TestPoolSubmitMany(t *testing.T) {
	pool := NewPool(3)
	ctx := context.Background()

	taskCount := 5
	var counter int64
	tasks := make([]func() error, taskCount)

	for i := 0; i < taskCount; i++ {
		tasks[i] = func() error {
			atomic.AddInt64(&counter, 1)
			return nil
		}
	}

	submitted, err := pool.SubmitMany(ctx, tasks)
	if err != nil {
		t.Fatalf("SubmitMany failed: %v", err)
	}
	if submitted != taskCount {
		t.Errorf("submitted = %d, want %d", submitted, taskCount)
	}

	pool.Wait()

	if counter != int64(taskCount) {
		t.Errorf("counter = %d, want %d", counter, taskCount)
	}
}

func TestPoolSubmitManyWithCancellation(t *testing.T) {
	pool := NewPool(1)
	ctx, cancel := context.WithCancel(context.Background())

	// Block the pool with a long-running task
	_ = pool.Submit(ctx, func() error {
		time.Sleep(time.Second * 2)
		return nil
	})

	// Cancel context
	cancel()

	// Try to submit more tasks
	tasks := []func() error{
		func() error { return nil },
		func() error { return nil },
	}

	submitted, err := pool.SubmitMany(ctx, tasks)
	if err == nil {
		t.Error("SubmitMany should fail with canceled context")
	}
	if submitted != 0 {
		t.Errorf("submitted = %d, want 0", submitted)
	}
}

func TestPoolSubmitFunc(t *testing.T) {
	pool := NewPool(2)
	ctx := context.Background()

	executed := false
	fn := func() {
		executed = true
	}

	err := pool.SubmitFunc(ctx, fn)
	if err != nil {
		t.Fatalf("SubmitFunc failed: %v", err)
	}

	pool.Wait()

	if !executed {
		t.Error("function was not executed")
	}
}

func TestPoolRaceConditions(t *testing.T) {
	pool := NewPool(10)
	ctx := context.Background()

	var counter int64
	taskCount := 100

	var wg sync.WaitGroup
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		go func() {
			defer wg.Done()
			err := pool.Submit(ctx, func() error {
				atomic.AddInt64(&counter, 1)
				return nil
			})
			if err != nil {
				t.Errorf("Submit failed: %v", err)
			}
		}()
	}

	wg.Wait()
	pool.Wait()

	if counter != int64(taskCount) {
		t.Errorf("counter = %d, want %d", counter, taskCount)
	}
}

// Benchmark tests

func BenchmarkPoolSubmit(b *testing.B) {
	pool := NewPool(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.Submit(ctx, func() error {
			return nil
		})
	}
	pool.Wait()
}

func BenchmarkPoolSubmitParallel(b *testing.B) {
	pool := NewPool(10)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = pool.Submit(ctx, func() error {
				return nil
			})
		}
	})
	pool.Wait()
}

func BenchmarkPoolWithWork(b *testing.B) {
	pool := NewPool(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.Submit(ctx, func() error {
			// Simulate some work
			time.Sleep(time.Microsecond)
			return nil
		})
	}
	pool.Wait()
}
