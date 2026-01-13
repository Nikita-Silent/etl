// Package logger provides structured logging for the application.
// Uses Go 1.21+ slog package for structured, leveled logging.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level represents log level
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger wraps slog.Logger with convenience methods
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output io.Writer
}

// New creates a new Logger with the given configuration
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.000")),
				}
			}
			return a
		},
	}

	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// Default creates a logger with default settings
func Default() *Logger {
	return New(Config{
		Level:  "info",
		Format: "text",
		Output: os.Stdout,
	})
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithRequestID returns a logger with request ID context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.With("request_id", requestID),
	}
}

// WithComponent returns a logger with component context
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.With("component", component),
	}
}

// WithKassa returns a logger with kassa context
func (l *Logger) WithKassa(kassaCode, folderName string) *Logger {
	return &Logger{
		Logger: l.With("kassa_code", kassaCode, "folder", folderName),
	}
}

// ETL logging helpers

// LogETLStart logs ETL pipeline start
func (l *Logger) LogETLStart(ctx context.Context, date string) {
	l.InfoContext(ctx, "ETL pipeline started",
		"date", date,
		"event", "etl_start",
	)
}

// LogETLEnd logs ETL pipeline completion
func (l *Logger) LogETLEnd(ctx context.Context, date string, filesProcessed, transactionsLoaded int, err error) {
	if err != nil {
		l.ErrorContext(ctx, "ETL pipeline failed",
			"date", date,
			"event", "etl_end",
			"files_processed", filesProcessed,
			"transactions_loaded", transactionsLoaded,
			"error", err.Error(),
		)
	} else {
		l.InfoContext(ctx, "ETL pipeline completed",
			"date", date,
			"event", "etl_end",
			"files_processed", filesProcessed,
			"transactions_loaded", transactionsLoaded,
		)
	}
}

// LogFileProcessed logs file processing
func (l *Logger) LogFileProcessed(ctx context.Context, filePath string, transactions int, err error) {
	if err != nil {
		l.ErrorContext(ctx, "File processing failed",
			"file", filePath,
			"event", "file_processed",
			"error", err.Error(),
		)
	} else {
		l.DebugContext(ctx, "File processed",
			"file", filePath,
			"event", "file_processed",
			"transactions", transactions,
		)
	}
}

// LogDBOperation logs database operations
func (l *Logger) LogDBOperation(ctx context.Context, operation, table string, rowsAffected int, err error) {
	if err != nil {
		l.ErrorContext(ctx, "Database operation failed",
			"operation", operation,
			"table", table,
			"event", "db_operation",
			"error", err.Error(),
		)
	} else {
		l.DebugContext(ctx, "Database operation completed",
			"operation", operation,
			"table", table,
			"event", "db_operation",
			"rows_affected", rowsAffected,
		)
	}
}

// LogFTPOperation logs FTP operations
func (l *Logger) LogFTPOperation(ctx context.Context, operation, path string, err error) {
	if err != nil {
		l.ErrorContext(ctx, "FTP operation failed",
			"operation", operation,
			"path", path,
			"event", "ftp_operation",
			"error", err.Error(),
		)
	} else {
		l.DebugContext(ctx, "FTP operation completed",
			"operation", operation,
			"path", path,
			"event", "ftp_operation",
		)
	}
}
