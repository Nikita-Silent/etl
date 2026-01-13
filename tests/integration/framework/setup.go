//go:build integration
// +build integration

package framework

import (
	"context"
	"os"
	"testing"
)

// TestEnvironment wraps all test containers and helpers
type TestEnvironment struct {
	Postgres *PostgresContainer
	FTP      *FTPContainer
	Builder  *TestDataBuilder
	ctx      context.Context
}

// SetupTestEnvironment creates and starts all required test containers
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	postgres, err := NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Run migrations
	if err := postgres.RunMigrations(ctx); err != nil {
		postgres.Close(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Start FTP container
	ftp, err := NewFTPContainer(ctx)
	if err != nil {
		postgres.Close(ctx)
		t.Fatalf("Failed to start FTP container: %v", err)
	}

	// Setup FTP folder structure
	if err := ftp.SetupFolderStructure(ctx); err != nil {
		postgres.Close(ctx)
		ftp.Close(ctx)
		t.Fatalf("Failed to setup FTP folders: %v", err)
	}

	// Create test data builder
	builder := NewTestDataBuilder(postgres, ftp)

	env := &TestEnvironment{
		Postgres: postgres,
		FTP:      ftp,
		Builder:  builder,
		ctx:      ctx,
	}

	// Register cleanup
	t.Cleanup(func() {
		env.Teardown()
	})

	return env
}

// Teardown cleans up the test environment
func (env *TestEnvironment) Teardown() {
	if env.FTP != nil {
		env.FTP.Close(env.ctx)
	}
	if env.Postgres != nil {
		env.Postgres.Close(env.ctx)
	}
}

// Reset resets the test environment to a clean state
func (env *TestEnvironment) Reset(t *testing.T) {
	t.Helper()

	// Truncate database tables
	if err := env.Postgres.Truncate(env.ctx); err != nil {
		t.Fatalf("Failed to truncate database: %v", err)
	}

	// Clean FTP folders
	if err := env.FTP.CleanFolders(env.ctx); err != nil {
		t.Fatalf("Failed to clean FTP folders: %v", err)
	}
}

// LoadBasicData loads basic test data
func (env *TestEnvironment) LoadBasicData(t *testing.T) {
	t.Helper()

	dataset := GetBasicDataSet()
	if err := env.Builder.LoadSeedData(env.ctx, dataset); err != nil {
		t.Fatalf("Failed to load basic data: %v", err)
	}
}

// GetContext returns the test context
func (env *TestEnvironment) GetContext() context.Context {
	return env.ctx
}
