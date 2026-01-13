package logger

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "default text format",
			config: Config{
				Level:  "info",
				Format: "text",
			},
		},
		{
			name: "json format",
			config: Config{
				Level:  "debug",
				Format: "json",
			},
		},
		{
			name:   "empty config uses defaults",
			config: Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tt.config.Output = buf

			log := New(tt.config)

			if log == nil {
				t.Fatal("New() returned nil")
			}

			log.Info("test message")

			if buf.Len() == 0 {
				t.Error("Expected log output, got none")
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo}, // defaults to info
		{"", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLogger_WithRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text"})

	logWithID := log.WithRequestID("req_123")
	logWithID.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "req_123") {
		t.Errorf("Expected output to contain request_id, got: %s", output)
	}
}

func TestLogger_WithComponent(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text"})

	logWithComp := log.WithComponent("parser")
	logWithComp.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "parser") {
		t.Errorf("Expected output to contain component, got: %s", output)
	}
}

func TestLogger_WithKassa(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text"})

	logWithKassa := log.WithKassa("001", "folder1")
	logWithKassa.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "001") {
		t.Errorf("Expected output to contain kassa_code, got: %s", output)
	}
	if !strings.Contains(output, "folder1") {
		t.Errorf("Expected output to contain folder, got: %s", output)
	}
}

func TestLogger_LogETLStart(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text"})

	log.LogETLStart(context.Background(), "2024-12-01")

	output := buf.String()
	if !strings.Contains(output, "ETL pipeline started") {
		t.Errorf("Expected 'ETL pipeline started', got: %s", output)
	}
	if !strings.Contains(output, "2024-12-01") {
		t.Errorf("Expected date in output, got: %s", output)
	}
}

func TestLogger_LogETLEnd_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text"})

	log.LogETLEnd(context.Background(), "2024-12-01", 10, 1000, nil)

	output := buf.String()
	if !strings.Contains(output, "completed") {
		t.Errorf("Expected 'completed', got: %s", output)
	}
}

func TestLogger_LogETLEnd_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text"})

	log.LogETLEnd(context.Background(), "2024-12-01", 5, 500, errors.New("test error"))

	output := buf.String()
	if !strings.Contains(output, "failed") {
		t.Errorf("Expected 'failed', got: %s", output)
	}
	if !strings.Contains(output, "test error") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestLogger_LogFileProcessed(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text", Level: "debug"})

	log.LogFileProcessed(context.Background(), "/path/to/file.txt", 100, nil)

	output := buf.String()
	if !strings.Contains(output, "file.txt") {
		t.Errorf("Expected file path in output, got: %s", output)
	}
}

func TestLogger_LogDBOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text", Level: "debug"})

	log.LogDBOperation(context.Background(), "insert", "transactions", 100, nil)

	output := buf.String()
	if !strings.Contains(output, "transactions") {
		t.Errorf("Expected table name in output, got: %s", output)
	}
}

func TestLogger_LogFTPOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text", Level: "debug"})

	log.LogFTPOperation(context.Background(), "download", "/remote/path", nil)

	output := buf.String()
	if !strings.Contains(output, "download") {
		t.Errorf("Expected operation in output, got: %s", output)
	}
}

func TestDefault(t *testing.T) {
	log := Default()
	if log == nil {
		t.Fatal("Default() returned nil")
	}
}

// Benchmarks
func BenchmarkLogger_Info(b *testing.B) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "text"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		log.Info("benchmark message", "key", "value")
	}
}

func BenchmarkLogger_JSON(b *testing.B) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		log.Info("benchmark message", "key", "value")
	}
}
