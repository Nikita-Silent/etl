package migrate

import (
	"testing"

	"github.com/user/go-frontol-loader/pkg/models"
)

func TestNewMigrator_InvalidConfig(t *testing.T) {
	cfg := &models.Config{
		DBHost:     "invalid_host_that_does_not_exist",
		DBPort:     5432,
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "test",
		DBSSLMode:  "disable",
	}

	_, err := NewMigrator(cfg)
	if err == nil {
		t.Error("Expected error for invalid database host, got nil")
	}
}

func TestStatus_Fields(t *testing.T) {
	status := Status{
		Version: 3,
		Dirty:   false,
		Error:   nil,
	}

	if status.Version != 3 {
		t.Errorf("Status.Version = %v, want 3", status.Version)
	}
	if status.Dirty {
		t.Error("Status.Dirty should be false")
	}
	if status.Error != nil {
		t.Errorf("Status.Error should be nil, got %v", status.Error)
	}
}

// TestMigrationsEmbedded verifies that migrations are properly embedded
func TestMigrationsEmbedded(t *testing.T) {
	// Check that embedded filesystem contains migrations
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		t.Fatalf("Failed to read embedded migrations directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("No migrations found in embedded filesystem")
	}

	// Verify expected migration files exist
	expectedFiles := []string{
		"000001_init_schema.up.sql",
		"000001_init_schema.down.sql",
	}

	fileMap := make(map[string]bool)
	for _, entry := range entries {
		fileMap[entry.Name()] = true
	}

	for _, expected := range expectedFiles {
		if !fileMap[expected] {
			t.Errorf("Expected migration file not found: %s", expected)
		}
	}
}

// TestMigrationFileContent verifies migration files have content
func TestMigrationFileContent(t *testing.T) {
	files := []string{
		"migrations/000001_init_schema.up.sql",
		"migrations/000001_init_schema.down.sql",
	}

	for _, file := range files {
		content, err := migrationsFS.ReadFile(file)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Migration file %s is empty", file)
		}
	}
}
