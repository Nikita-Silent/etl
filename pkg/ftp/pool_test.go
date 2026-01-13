package ftp

import (
	"sync"
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
)

func TestPoolGetPut(t *testing.T) {
	// This is a basic test of pool mechanics
	// For full integration test with real FTP, use integration tests
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	// Create a small pool for testing
	// Note: This will fail if no FTP server is available
	// In a real scenario, you'd use a mock FTP server or skip this test
	pool, err := NewPool(cfg, 2)
	if err != nil {
		t.Skipf("Skipping test - FTP server not available: %v", err)
		return
	}
	defer pool.Close()

	// Test Get
	client1, err := pool.Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if client1 == nil {
		t.Fatal("Get() returned nil client")
	}

	// Test Put
	err = pool.Put(client1)
	if err != nil {
		t.Errorf("Put() failed: %v", err)
	}

	// Get again to verify it was returned to pool
	client2, err := pool.Get()
	if err != nil {
		t.Fatalf("Get() after Put() failed: %v", err)
	}
	if client2 == nil {
		t.Fatal("Get() after Put() returned nil")
	}

	pool.Put(client2)
}

func TestPoolConcurrency(t *testing.T) {
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	poolSize := 3
	pool, err := NewPool(cfg, poolSize)
	if err != nil {
		t.Skipf("Skipping test - FTP server not available: %v", err)
		return
	}
	defer pool.Close()

	// Test concurrent Get/Put operations
	var wg sync.WaitGroup
	goroutines := 10
	iterations := 5

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				client, err := pool.Get()
				if err != nil {
					t.Errorf("Goroutine %d: Get() failed: %v", id, err)
					return
				}

				// Simulate some work
				time.Sleep(time.Millisecond * 10)

				err = pool.Put(client)
				if err != nil {
					t.Errorf("Goroutine %d: Put() failed: %v", id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestPoolClose(t *testing.T) {
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	pool, err := NewPool(cfg, 2)
	if err != nil {
		t.Skipf("Skipping test - FTP server not available: %v", err)
		return
	}

	// Close pool
	err = pool.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verify Get() fails after close
	_, err = pool.Get()
	if err == nil {
		t.Error("Get() should fail after Close()")
	}

	// Verify double close doesn't panic
	err = pool.Close()
	if err != nil {
		t.Errorf("Second Close() failed: %v", err)
	}
}

func TestPoolWithConnection(t *testing.T) {
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	pool, err := NewPool(cfg, 2)
	if err != nil {
		t.Skipf("Skipping test - FTP server not available: %v", err)
		return
	}
	defer pool.Close()

	// Test WithConnection executes function and returns connection
	executed := false
	err = pool.WithConnection(func(client *Client) error {
		if client == nil {
			t.Error("WithConnection passed nil client")
		}
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("WithConnection() failed: %v", err)
	}
	if !executed {
		t.Error("WithConnection() did not execute function")
	}

	// Verify pool still has connections available after WithConnection
	client, err := pool.Get()
	if err != nil {
		t.Errorf("Get() after WithConnection() failed: %v", err)
	}
	if client == nil {
		t.Error("Pool has no connections after WithConnection()")
	}
	pool.Put(client)
}

func TestPoolZeroSize(t *testing.T) {
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	// Test with size 0 - should default to 5
	pool, err := NewPool(cfg, 0)
	if err != nil {
		t.Skipf("Skipping test - FTP server not available: %v", err)
		return
	}
	defer pool.Close()

	// Should be able to get at least 1 connection
	client, err := pool.Get()
	if err != nil {
		t.Fatalf("Get() failed with default pool size: %v", err)
	}
	if client == nil {
		t.Fatal("Get() returned nil with default pool size")
	}
	pool.Put(client)
}

func TestPoolInterfaceImplementation(t *testing.T) {
	// Compile-time check that Pool implements FTPClient
	var _ FTPClient = (*Pool)(nil)

	// Test that we can use Pool as FTPClient
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	pool, err := NewPool(cfg, 1)
	if err != nil {
		t.Skipf("Skipping test - FTP server not available: %v", err)
		return
	}
	defer pool.Close()
}

func TestPoolConcurrentWithConnection(t *testing.T) {
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	poolSize := 3
	pool, err := NewPool(cfg, poolSize)
	if err != nil {
		t.Skipf("Skipping test - FTP server not available: %v", err)
		return
	}
	defer pool.Close()

	// Test concurrent WithConnection calls
	var wg sync.WaitGroup
	goroutines := 10
	successCount := 0
	var countMutex sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			err := pool.WithConnection(func(client *Client) error {
				// Simulate some work
				time.Sleep(time.Millisecond * 10)
				return nil
			})

			if err == nil {
				countMutex.Lock()
				successCount++
				countMutex.Unlock()
			} else {
				t.Errorf("Goroutine %d: WithConnection() failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	if successCount != goroutines {
		t.Errorf("Expected %d successful WithConnection calls, got %d", goroutines, successCount)
	}
}

// Benchmark tests

func BenchmarkPoolGetPut(b *testing.B) {
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	pool, err := NewPool(cfg, 5)
	if err != nil {
		b.Skipf("Skipping benchmark - FTP server not available: %v", err)
		return
	}
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client, err := pool.Get()
			if err != nil {
				b.Errorf("Get() failed: %v", err)
				return
			}
			pool.Put(client)
		}
	})
}

func BenchmarkPoolWithConnection(b *testing.B) {
	cfg := &models.Config{
		FTPHost:        "localhost",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{},
		LocalDir:       "/tmp/test",
	}

	pool, err := NewPool(cfg, 5)
	if err != nil {
		b.Skipf("Skipping benchmark - FTP server not available: %v", err)
		return
	}
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := pool.WithConnection(func(client *Client) error {
				return nil
			})
			if err != nil {
				b.Errorf("WithConnection() failed: %v", err)
				return
			}
		}
	})
}
