//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/db"
)

// TestDatabaseConnection tests database connection
func TestDatabaseConnection(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	cfg := getTestDBConfig()
	if cfg.DBPassword == "" {
		t.Skip("Skipping test - database password not set")
	}

	pool, err := db.NewPool(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

// TestDatabaseTransaction tests database transaction
func TestDatabaseTransaction(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	cfg := getTestDBConfig()
	if cfg.DBPassword == "" {
		t.Skip("Skipping test - database password not set")
	}

	pool, err := db.NewPool(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := pool.BeginTx(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Test that transaction can be rolled back
	if err := tx.Rollback(ctx); err != nil {
		t.Errorf("Failed to rollback transaction: %v", err)
	}
}
