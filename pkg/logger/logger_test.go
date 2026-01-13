package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	log := New(Config{Output: buf, Format: "json"})

	logWithID := log.WithRequestID("req_123")
	logWithID.Info("test message")

	payload := parseJSONLine(t, buf)
	if got := payload["request_id"]; got != "req_123" {
		t.Errorf("Expected request_id 'req_123', got: %v", got)
	}
}

func TestLogger_WithComponent(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json"})

	logWithComp := log.WithComponent("parser")
	logWithComp.Info("test message")

	payload := parseJSONLine(t, buf)
	if got := payload["component"]; got != "parser" {
		t.Errorf("Expected component 'parser', got: %v", got)
	}
}

func TestLogger_WithKassa(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json"})

	logWithKassa := log.WithKassa("001", "folder1")
	logWithKassa.Info("test message")

	payload := parseJSONLine(t, buf)
	if got := payload["kassa_code"]; got != "001" {
		t.Errorf("Expected kassa_code '001', got: %v", got)
	}
	if got := payload["folder"]; got != "folder1" {
		t.Errorf("Expected folder 'folder1', got: %v", got)
	}
}

func TestLogger_LogETLStart(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json"})

	log.LogETLStart(context.Background(), "2024-12-01")

	payload := parseJSONLine(t, buf)
	if payload["event"] != "etl_start" {
		t.Errorf("Expected event 'etl_start', got: %v", payload["event"])
	}
	if payload["date"] != "2024-12-01" {
		t.Errorf("Expected date '2024-12-01', got: %v", payload["date"])
	}
}

func TestLogger_LogETLEnd_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json"})

	log.LogETLEnd(context.Background(), "2024-12-01", 10, 1000, nil)

	payload := parseJSONLine(t, buf)
	if payload["event"] != "etl_end" {
		t.Errorf("Expected event 'etl_end', got: %v", payload["event"])
	}
	if payload["files_processed"] != float64(10) {
		t.Errorf("Expected files_processed 10, got: %v", payload["files_processed"])
	}
}

func TestLogger_LogETLEnd_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json"})

	log.LogETLEnd(context.Background(), "2024-12-01", 5, 500, errors.New("test error"))

	payload := parseJSONLine(t, buf)
	if payload["event"] != "etl_end" {
		t.Errorf("Expected event 'etl_end', got: %v", payload["event"])
	}
	if payload["error"] != "test error" {
		t.Errorf("Expected error 'test error', got: %v", payload["error"])
	}
}

func TestLogger_LogFileProcessed(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json", Level: "debug"})

	log.LogFileProcessed(context.Background(), "/path/to/file.txt", 100, nil)

	payload := parseJSONLine(t, buf)
	if payload["file"] != "/path/to/file.txt" {
		t.Errorf("Expected file path in output, got: %v", payload["file"])
	}
}

func TestLogger_LogDBOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json", Level: "debug"})

	log.LogDBOperation(context.Background(), "insert", "transactions", 100, nil)

	payload := parseJSONLine(t, buf)
	if payload["table"] != "transactions" {
		t.Errorf("Expected table name 'transactions', got: %v", payload["table"])
	}
}

func TestLogger_LogFTPOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{Output: buf, Format: "json", Level: "debug"})

	log.LogFTPOperation(context.Background(), "download", "/remote/path", nil)

	payload := parseJSONLine(t, buf)
	if payload["operation"] != "download" {
		t.Errorf("Expected operation 'download', got: %v", payload["operation"])
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

func TestNew_WithSlogBackend(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{
		Output:  buf,
		Format:  "text",
		Backend: "slog",
	})

	log.Info("test message", "key", "value")
	if buf.Len() == 0 {
		t.Fatal("Expected slog backend to emit output")
	}
}

func TestFilteringWriter_AllowsConfiguredEvents(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{
		Output:      buf,
		Format:      "json",
		FilterField: "event",
		FilterAllow: []string{"keep"},
	})

	log.Info("allowed log", "event", "keep")
	if buf.Len() == 0 {
		t.Fatal("Expected allowed log to be written")
	}

	buf.Reset()
	log.Info("dropped log", "event", "drop")
	if buf.Len() != 0 {
		t.Fatalf("Expected log to be filtered out, got: %s", buf.String())
	}
}

func TestFilteringWriter_PassThroughWhenNoAllowlist(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{
		Output:      buf,
		Format:      "json",
		FilterField: "event",
	})

	log.Info("log without filter", "event", "anything")
	if buf.Len() == 0 {
		t.Fatal("Expected log to be written when allowlist is empty")
	}
}

func TestFilteringWriter_UsesEnvWhenNotProvided(t *testing.T) {
	buf := &bytes.Buffer{}
	t.Setenv("LOG_FILTER_FIELD", "event")
	t.Setenv("LOG_FILTER_ALLOW", "allowed,other")

	log := New(Config{
		Output: buf,
		Format: "json",
	})

	log.Info("allowed", "event", "allowed")
	if buf.Len() == 0 {
		t.Fatal("Expected log to be written when event is allowed from env")
	}

	buf.Reset()
	log.Info("blocked", "event", "blocked")
	if buf.Len() != 0 {
		t.Fatalf("Expected log to be filtered by env allowlist, got: %s", buf.String())
	}
}

func parseJSONLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	line := bytes.TrimSpace(buf.Bytes())
	if len(line) == 0 {
		t.Fatal("expected log output, got empty buffer")
	}

	result := make(map[string]any)
	if err := json.Unmarshal(line, &result); err != nil {
		t.Fatalf("failed to parse log line as JSON: %v; content: %s", err, string(line))
	}

	return result
}
