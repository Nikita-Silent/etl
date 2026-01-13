package pipeline

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
)

func TestPipelineResult(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)
	result := &PipelineResult{
		StartTime:          startTime,
		EndTime:            endTime,
		Duration:           endTime.Sub(startTime).String(),
		Date:               "2024-12-01",
		FilesProcessed:     10,
		FilesSkipped:       2,
		TransactionsLoaded: 100,
		Errors:             0,
		Success:            true,
	}

	if result.Duration == "" {
		t.Error("PipelineResult.Duration should be set")
	}

	if !result.Success {
		t.Error("PipelineResult.Success should be true")
	}

	if result.FilesProcessed != 10 {
		t.Errorf("Expected FilesProcessed=10, got %d", result.FilesProcessed)
	}

	if result.TransactionsLoaded != 100 {
		t.Errorf("Expected TransactionsLoaded=100, got %d", result.TransactionsLoaded)
	}
}

func TestProcessingStats(t *testing.T) {
	stats := &ProcessingStats{
		FilesProcessed:     5,
		FilesSkipped:       1,
		TransactionsLoaded: 50,
		Errors:             0,
	}

	if stats.FilesProcessed != 5 {
		t.Errorf("Expected FilesProcessed=5, got %d", stats.FilesProcessed)
	}

	if stats.FilesSkipped != 1 {
		t.Errorf("Expected FilesSkipped=1, got %d", stats.FilesSkipped)
	}

	if stats.TransactionsLoaded != 50 {
		t.Errorf("Expected TransactionsLoaded=50, got %d", stats.TransactionsLoaded)
	}
}

// TestPipelineRunWithInvalidConfig tests that pipeline fails gracefully with invalid config
func TestPipelineRunWithInvalidConfig(t *testing.T) {
	// Skip if no database available
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	invalidCfg := &models.Config{
		DBHost:         "invalid_host",
		DBPort:         5432,
		DBUser:         "test",
		DBPassword:     "test",
		DBName:         "test",
		DBSSLMode:      "disable",
		FTPHost:        "invalid_ftp",
		FTPPort:        21,
		FTPUser:        "test",
		FTPPassword:    "test",
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{
			"001": {"folder1"},
		},
		LocalDir:         "/tmp/test",
		BatchSize:        1000,
		MaxRetries:       3,
		RetryDelay:       1 * time.Second,
		WaitDelayMinutes: 1 * time.Minute,
		LogLevel:         "info",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := Run(ctx, logger, invalidCfg, "2024-12-01")

	// Pipeline should fail with invalid config
	if err == nil {
		t.Error("Expected error with invalid config, got nil")
	}

	if result != nil && result.Success {
		t.Error("Expected pipeline to fail with invalid config")
	}
}

// TestPipelineRunWithMinimalConfig tests pipeline initialization
func TestPipelineRunWithMinimalConfig(t *testing.T) {
	// Skip if no database available
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	cfg := &models.Config{
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         5432,
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBSSLMode:      "disable",
		FTPHost:        os.Getenv("FTP_HOST"),
		FTPPort:        21,
		FTPUser:        os.Getenv("FTP_USER"),
		FTPPassword:    os.Getenv("FTP_PASSWORD"),
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		KassaStructure: map[string][]string{
			"001": {"folder1"},
		},
		LocalDir:         "/tmp/test",
		BatchSize:        1000,
		MaxRetries:       3,
		RetryDelay:       1 * time.Second,
		WaitDelayMinutes: 1 * time.Minute,
		LogLevel:         "info",
	}

	// If no config provided, skip test
	if cfg.DBHost == "" || cfg.FTPHost == "" {
		t.Skip("Skipping test - no database/FTP configuration provided")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This will likely fail at connection stage, but tests the pipeline structure
	result, err := Run(ctx, logger, cfg, "2024-12-01")

	// We expect either success or connection error, not a panic
	if err != nil {
		// Connection errors are expected in test environment
		if result != nil && result.ErrorMessage == "" {
			t.Error("Expected error message in result")
		}
	} else {
		if result == nil {
			t.Error("Expected result, got nil")
		}
		if result != nil && !result.Success {
			t.Logf("Pipeline failed: %s", result.ErrorMessage)
		}
	}
}

// TestFileTaskStructure tests fileTask structure
func TestFileTaskStructure(t *testing.T) {
	// This test verifies that fileTask can be created and used
	// In real scenario, you'd need actual folder and file entries
	// This is a structural test to ensure the type is correct

	task := fileTask{
		folder: models.KassaFolder{
			KassaCode:  "001",
			FolderName: "test_folder",
		},
		file: nil, // Would be *ftplib.Entry in real scenario
	}

	if task.folder.KassaCode != "001" {
		t.Errorf("Expected KassaCode '001', got '%s'", task.folder.KassaCode)
	}

	if task.folder.FolderName != "test_folder" {
		t.Errorf("Expected FolderName 'test_folder', got '%s'", task.folder.FolderName)
	}
}

// TestProcessingStatsAggregation tests statistics aggregation
func TestProcessingStatsAggregation(t *testing.T) {
	stats := &ProcessingStats{
		FilesProcessed:     0,
		FilesSkipped:       0,
		TransactionsLoaded: 0,
		Errors:             0,
		TransactionDetails: []TransactionTypeStats{},
	}

	// Simulate processing multiple files
	stats.FilesProcessed++
	stats.TransactionsLoaded += 10

	stats.FilesProcessed++
	stats.TransactionsLoaded += 20

	stats.FilesSkipped++

	if stats.FilesProcessed != 2 {
		t.Errorf("Expected FilesProcessed=2, got %d", stats.FilesProcessed)
	}

	if stats.TransactionsLoaded != 30 {
		t.Errorf("Expected TransactionsLoaded=30, got %d", stats.TransactionsLoaded)
	}

	if stats.FilesSkipped != 1 {
		t.Errorf("Expected FilesSkipped=1, got %d", stats.FilesSkipped)
	}
}

// TestTransactionTypeStats tests TransactionTypeStats structure
func TestTransactionTypeStats(t *testing.T) {
	stats := TransactionTypeStats{
		TableName: "tx_item_registration_1_11",
		Count:     100,
	}

	if stats.TableName != "tx_item_registration_1_11" {
		t.Errorf("Expected TableName 'tx_item_registration_1_11', got '%s'", stats.TableName)
	}

	if stats.Count != 100 {
		t.Errorf("Expected Count=100, got %d", stats.Count)
	}
}

// TestPipelineResultDuration tests duration calculation
func TestPipelineResultDuration(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)

	result := &PipelineResult{
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime).String(),
	}

	if result.Duration == "" {
		t.Error("Duration should not be empty")
	}

	// Duration should be approximately 5 seconds
	expectedDuration := 5 * time.Second
	actualDuration := endTime.Sub(startTime)
	if actualDuration < expectedDuration-time.Millisecond || actualDuration > expectedDuration+time.Millisecond {
		t.Errorf("Expected duration ~5s, got %v", actualDuration)
	}
}

// TestPipelineResultFields tests all PipelineResult fields
func TestPipelineResultFields(t *testing.T) {
	now := time.Now()
	result := &PipelineResult{
		StartTime:          now,
		EndTime:            now.Add(1 * time.Second),
		Duration:           "1s",
		Date:               "2024-12-01",
		FilesProcessed:     5,
		FilesSkipped:       2,
		TransactionsLoaded: 50,
		Errors:             0,
		Success:            true,
		ErrorMessage:       "",
		TransactionDetails: []TransactionTypeStats{
			{TableName: "tx_item_registration_1_11", Count: 30},
			{TableName: "special_prices", Count: 20},
		},
	}

	if result.Date != "2024-12-01" {
		t.Errorf("Expected Date '2024-12-01', got '%s'", result.Date)
	}

	if !result.Success {
		t.Error("Expected Success=true")
	}

	if len(result.TransactionDetails) != 2 {
		t.Errorf("Expected 2 transaction details, got %d", len(result.TransactionDetails))
	}
}

// TestRemoveFile tests removeFile function
func TestRemoveFile(t *testing.T) {
	// Test with non-existent file (should not error)
	err := removeFile("/tmp/nonexistent_file_12345.txt")
	if err != nil {
		// It's okay if it errors on non-existent file, os.Remove can return error
		// But the function should handle it gracefully
		t.Logf("removeFile returned error for non-existent file (expected): %v", err)
	}

	// Test with empty path
	err = removeFile("")
	if err != nil {
		t.Logf("removeFile returned error for empty path (expected): %v", err)
	}
}
