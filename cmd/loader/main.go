package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/operations"
	"github.com/user/go-frontol-loader/pkg/pipeline"
)

func firstIssueStage(result *pipeline.PipelineResult) string {
	if result == nil || len(result.ErrorSamples) == 0 {
		return ""
	}
	return result.ErrorSamples[0].Stage
}

func main() {
	defaultLogger := logger.New(logger.Config{
		Level:   "info",
		Format:  os.Getenv("LOG_FORMAT"),
		Output:  os.Stdout,
		Backend: os.Getenv("LOG_BACKEND"),
	})
	defer func() { _ = defaultLogger.Close() }()
	slog.SetDefault(defaultLogger.Logger)

	// Parse command line arguments
	var date string
	if len(os.Args) > 1 {
		date = os.Args[1]
		// Validate date format
		if _, err := time.Parse("2006-01-02", date); err != nil {
			// #nosec G706 -- invalid CLI date is logged for operator troubleshooting.
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
		Format:  cfg.LogFormat,
		Output:  os.Stdout,
		Backend: cfg.LogBackend,
	})
	defer func() { _ = loggerInstance.Close() }()
	operationID := logger.NewOperationID()
	log := loggerInstance.WithComponent("loader").WithOperationID(operationID)
	opStore := operations.NewStore(cfg, loggerInstance)
	defer opStore.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.EffectiveCLIRunTimeout())
	defer cancel()

	log.InfoContext(ctx, "Timeout configuration loaded",
		"event", "timeout_config_loaded",
		"cli_run_timeout", cfg.EffectiveCLIRunTimeout(),
		"db_connect_timeout", cfg.EffectiveDBConnectTimeout(),
		"ftp_connect_timeout", cfg.EffectiveFTPConnectTimeout(),
		"pipeline_load_timeout", cfg.EffectivePipelineLoadTimeout(),
		"wait_delay", cfg.WaitDelayMinutes,
	)
	_ = opStore.Start(ctx, operations.Record{
		OperationID:   operationID,
		OperationType: "cli_load",
		Status:        operations.StatusProcessing,
		Date:          date,
		Component:     "loader",
		StartedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	// Run ETL pipeline
	log.InfoContext(ctx, "Starting ETL pipeline",
		"date", date,
		"event", "etl_start",
	)
	result, err := pipeline.Run(ctx, log.Logger, cfg, date)
	if err != nil {
		now := time.Now()
		_ = opStore.Update(ctx, operations.Record{
			OperationID:   operationID,
			OperationType: "cli_load",
			Status:        operations.StatusFailed,
			Date:          date,
			Component:     "loader",
			UpdatedAt:     now,
			FinishedAt:    &now,
			ErrorMessage:  err.Error(),
			FailedStage:   "pipeline",
		})
		log.ErrorContext(ctx, "ETL pipeline failed",
			"error", err.Error(),
			"event", "etl_failed",
		)
		cancel()
		os.Exit(1)
	}
	now := time.Now()
	status := operations.StatusCompleted
	if result.Status == pipeline.PipelineStatusPartial {
		status = operations.StatusPartial
	}
	_ = opStore.Update(ctx, operations.Record{
		OperationID:   operationID,
		OperationType: "cli_load",
		Status:        status,
		Date:          date,
		Component:     "loader",
		UpdatedAt:     now,
		FinishedAt:    &now,
		ErrorMessage:  result.ErrorMessage,
		FailedStage:   firstIssueStage(result),
	})

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
