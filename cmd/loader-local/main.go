package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/parser"
	"github.com/user/go-frontol-loader/pkg/repository"
)

func main() {
	defaultLogger := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
		Output: os.Stdout,
	})
	slog.SetDefault(defaultLogger.Logger)

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file_path>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s /path/to/frontol_export.txt\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Create logger
	loggerInstance := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
		Output: os.Stdout,
	})
	log := loggerInstance.With("component", "loader-local")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	// Initialize database connection
	database, err := db.NewPool(cfg)
	if err != nil {
		slog.Error("Failed to connect to database",
			"error", err.Error(),
		)
		os.Exit(1)
	}
	defer func() {
		database.Close()
	}()

	// Initialize loader
	loader := repository.NewLoader(database)

	// Parse file
	log.Info("Parsing file",
		"file", filePath,
		"event", "file_parse_start",
	)
	startTime := time.Now()

	// Use "local" as source folder for local file processing
	transactions, header, err := parser.ParseFile(filePath, "local")
	if err != nil {
		slog.Error("Failed to parse file",
			"file", filePath,
			"error", err.Error(),
		)
		database.Close()
		os.Exit(1)
	}

	// Print file header information
	log.Info("File header",
		"processed", header.Processed,
		"db_id", header.DBID,
		"report_number", header.ReportNum,
		"event", "file_header",
	)

	// Check if file is already processed
	if header.Processed {
		log.Info("File is already processed, skipping",
			"event", "file_already_processed",
		)
		return
	}

	// Load data into database
	log.Info("Loading data into database",
		"event", "db_load_start",
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := loader.LoadFileData(ctx, transactions); err != nil {
		slog.Error("Failed to load data",
			"error", err.Error(),
		)
		cancel()
		database.Close()
		os.Exit(1)
	}

	// Print statistics
	loader.PrintStatistics(ctx, transactions, startTime)

	log.Info("Successfully processed file",
		"file", filePath,
		"event", "file_process_complete",
	)
}
