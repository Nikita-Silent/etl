package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/pipeline"
)

func main() {
	defaultLogger := logger.New(logger.Config{
		Level:   "info",
		Format:  "text",
		Output:  os.Stdout,
		Backend: os.Getenv("LOG_BACKEND"),
	})
	slog.SetDefault(defaultLogger.Logger)

	// Parse command line arguments
	var date string
	if len(os.Args) > 1 {
		date = os.Args[1]
		// Validate date format
		if _, err := time.Parse("2006-01-02", date); err != nil {
			slog.Error("Invalid date format",
				"date", date,
				"error", err.Error(),
			)
			os.Exit(1)
		}
	} else {
		// Use current date if not provided
		date = time.Now().Format("2006-01-02")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	// Create logger
	loggerInstance := logger.New(logger.Config{
		Level:   cfg.LogLevel,
		Format:  "text",
		Output:  os.Stdout,
		Backend: cfg.LogBackend,
	})
	log := loggerInstance.WithComponent("loader")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Run ETL pipeline
	log.InfoContext(ctx, "Starting ETL pipeline",
		"date", date,
		"event", "etl_start",
	)
	result, err := pipeline.Run(ctx, log.Logger, cfg, date)
	if err != nil {
		log.ErrorContext(ctx, "ETL pipeline failed",
			"error", err.Error(),
			"event", "etl_failed",
		)
		cancel()
		os.Exit(1)
	}

	// Print results
	log.InfoContext(ctx, "ETL pipeline completed successfully",
		"duration", result.Duration,
		"files_processed", result.FilesProcessed,
		"files_skipped", result.FilesSkipped,
		"transactions_loaded", result.TransactionsLoaded,
		"errors", result.Errors,
		"event", "etl_complete",
	)
}
