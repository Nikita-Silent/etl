//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/models"
	"github.com/user/go-frontol-loader/pkg/repository"
)

// TestLoaderWithDatabase tests Loader with real database
func TestLoaderWithDatabase(t *testing.T) {
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

	loader := repository.NewLoader(pool)
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}

	// Test GetTransactionCount with empty data
	count := loader.GetTransactionCount(map[string]interface{}{})
	if count != 0 {
		t.Errorf("Expected 0 transactions, got %d", count)
	}

	// Test GetTransactionDetails with empty data
	details := loader.GetTransactionDetails(map[string]interface{}{})
	if len(details) != 0 {
		t.Errorf("Expected 0 details, got %d", len(details))
	}
}

// TestLoaderPrintStatistics tests PrintStatistics with real database
func TestLoaderPrintStatistics(t *testing.T) {
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

	loader := repository.NewLoader(pool)
	ctx := context.Background()

	transactions := map[string]interface{}{
		"tx_item_registration_1_11": []models.TxItemRegistration1_11{
			{TransactionIDUnique: 1, SourceFolder: "test"},
		},
	}

	// Should not panic
	loader.PrintStatistics(ctx, transactions, time.Now())
}
