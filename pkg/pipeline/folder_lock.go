package pipeline

import (
	"context"
	"errors"
	"sync"
	"time"
)

var errFolderLockTimeout = errors.New("folder lock timeout")

type folderLockManager struct {
	mu    sync.Mutex
	locks map[string]struct{}
}

func newFolderLockManager() *folderLockManager {
	return &folderLockManager{locks: make(map[string]struct{})}
}

var defaultFolderLocks = newFolderLockManager()

func (m *folderLockManager) acquire(ctx context.Context, key string, retryDelay, timeout time.Duration) (func(), time.Duration, error) {
	if retryDelay <= 0 {
		retryDelay = 100 * time.Millisecond
	}
	startedAt := time.Now()
	deadline := startedAt.Add(timeout)

	for {
		m.mu.Lock()
		if _, exists := m.locks[key]; !exists {
			m.locks[key] = struct{}{}
			m.mu.Unlock()
			var once sync.Once
			return func() {
				once.Do(func() {
					m.mu.Lock()
					delete(m.locks, key)
					m.mu.Unlock()
				})
			}, time.Since(startedAt), nil
		}
		m.mu.Unlock()

		if timeout <= 0 || time.Now().After(deadline) {
			return nil, time.Since(startedAt), errFolderLockTimeout
		}

		waitFor := retryDelay
		if remaining := time.Until(deadline); remaining < waitFor {
			waitFor = remaining
		}

		timer := time.NewTimer(waitFor)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, time.Since(startedAt), ctx.Err()
		case <-timer.C:
		}
	}
}
