package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/auth"
	"github.com/user/go-frontol-loader/pkg/models"
)

func newTestServer(t *testing.T, token string) *Server {
	t.Helper()

	cfg := &models.Config{
		DBHost:             "invalid",
		DBPort:             5432,
		DBUser:             "user",
		DBPassword:         "pass",
		DBName:             "db",
		DBSSLMode:          "disable",
		FTPHost:            "invalid",
		FTPPort:            21,
		FTPUser:            "user",
		FTPPassword:        "pass",
		FTPRequestDir:      "/request",
		FTPResponseDir:     "/response",
		ServerPort:         8080,
		WebhookBearerToken: token,
	}

	s := NewServer(cfg)
	// Prevent worker start in tests by marking queues active.
	loadQueue := s.queueManager.GetOrCreateQueue(OperationTypeLoad)
	loadQueue.workerStarted = true
	downloadQueue := s.queueManager.GetOrCreateQueue(OperationTypeDownload)
	downloadQueue.workerStarted = true

	return s
}

func newTestMux(s *Server) *http.ServeMux {
	mux := http.NewServeMux()
	bearerAuth := auth.BearerAuthMiddleware(s.logger.Logger, s.config.WebhookBearerToken)

	mux.HandleFunc("/api/load", bearerAuth(s.webhookHandler))
	mux.HandleFunc("/api/files", bearerAuth(s.downloadHandler))
	mux.HandleFunc("/api/queue/status", bearerAuth(s.queueStatusHandler))
	mux.HandleFunc("/api/kassas", bearerAuth(s.listKassasHandler))
	mux.HandleFunc("/api/health", s.healthHandler)
	mux.HandleFunc("/api/docs", s.docsHandler)
	mux.HandleFunc("/api/openapi.yaml", s.openAPIHandler)

	return mux
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "api", "openapi.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatalf("repo root not found from %s", wd)
	return ""
}

func TestWebhookHandler_RequiresAuth(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodPost, "/api/load", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWebhookHandler_InvalidJSON(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodPost, "/api/load", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWebhookHandler_InvalidDate(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodPost, "/api/load", strings.NewReader(`{"date":"2024-13-01"}`))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWebhookHandler_Queued(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodPost, "/api/load", strings.NewReader(`{"date":"2024-12-01"}`))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["status"] != "queued" {
		t.Fatalf("expected status queued, got %v", payload["status"])
	}
	if payload["request_id"] == "" {
		t.Fatalf("expected request_id to be set")
	}
}

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodGet, "/api/load", nil)
	rec := httptest.NewRecorder()
	s.webhookHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestDownloadHandler_MissingParams(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestQueueStatusHandler_OK(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/queue/status", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if shuttingDown, ok := payload["is_shutting_down"].(bool); !ok || shuttingDown {
		t.Fatalf("expected is_shutting_down=false, got %v", payload["is_shutting_down"])
	}
}

func TestDocsHandler_OK(t *testing.T) {
	root := findRepoRoot(t)
	specPath := filepath.Join(root, "api", "openapi.yaml")
	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read openapi spec: %v", err)
	}

	originalSpec := openAPISpec
	openAPISpec = specBytes
	t.Cleanup(func() {
		openAPISpec = originalSpec
	})

	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	rec := httptest.NewRecorder()
	s.docsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected text/html content type, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestOpenAPIHandler_OK(t *testing.T) {
	root := findRepoRoot(t)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodGet, "/api/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	s.openAPIHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/yaml") {
		t.Fatalf("expected yaml content type, got %s", rec.Header().Get("Content-Type"))
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("openapi:")) {
		t.Fatalf("expected openapi spec in response")
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodPost, "/api/health", nil)
	rec := httptest.NewRecorder()
	s.healthHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestListKassas_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodPost, "/api/kassas", nil)
	rec := httptest.NewRecorder()
	s.listKassasHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestDownloadHandler_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodPost, "/api/files", nil)
	rec := httptest.NewRecorder()
	s.downloadHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestDownloadHandler_SynchronousFailure(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files?source_folder=P13&date=2024-12-01", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if s.queueManager.GetQueueSize(OperationTypeDownload) != 0 {
		t.Fatalf("expected download queue size 0, got %d", s.queueManager.GetQueueSize(OperationTypeDownload))
	}
}

func TestEnqueueRejectedDuringShutdown(t *testing.T) {
	s := newTestServer(t, "")
	s.stopping.Store(true)

	err := s.enqueue(&QueueItem{
		RequestID:     "req-test",
		Date:          "2024-12-01",
		OperationType: OperationTypeLoad,
		Logger:        s.logger,
		CreatedAt:     time.Now(),
	})
	if err == nil {
		t.Fatal("enqueue() expected shutdown error, got nil")
	}
}

func TestQueueStatusHandler_ShuttingDownFlag(t *testing.T) {
	s := newTestServer(t, "token")
	s.stopping.Store(true)
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/queue/status", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if shuttingDown, ok := payload["is_shutting_down"].(bool); !ok || !shuttingDown {
		t.Fatalf("expected is_shutting_down=true, got %v", payload["is_shutting_down"])
	}
}

func TestStopDrainsLoadQueue(t *testing.T) {
	s := NewServer(&models.Config{
		DBHost:             "invalid",
		DBPort:             5432,
		DBUser:             "user",
		DBPassword:         "pass",
		DBName:             "db",
		DBSSLMode:          "disable",
		FTPHost:            "invalid",
		FTPPort:            21,
		FTPUser:            "user",
		FTPPassword:        "pass",
		FTPRequestDir:      "/request",
		FTPResponseDir:     "/response",
		ServerPort:         8080,
		ShutdownTimeout:    time.Second,
		WaitDelayMinutes:   time.Millisecond,
		WebhookBearerToken: "token",
	})

	queue := s.queueManager.GetOrCreateQueue(OperationTypeLoad)
	queue.workerStarted = false

	for i := 0; i < 2; i++ {
		if err := s.enqueue(&QueueItem{
			RequestID:     "req-stop-" + time.Now().Format("150405.000000000"),
			Date:          "2024-12-01",
			OperationType: OperationTypeLoad,
			Logger:        s.logger,
			CreatedAt:     time.Now(),
		}); err != nil {
			t.Fatalf("enqueue() unexpected error: %v", err)
		}
	}

	completed := make(chan struct{})
	go func() {
		s.Stop()
		close(completed)
	}()

	select {
	case <-completed:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop() timed out waiting for queue drain")
	}

	if got := s.queueManager.GetQueueSize(OperationTypeLoad); got != 0 {
		t.Fatalf("expected load queue to drain, got %d items", got)
	}
}

func TestQueueStatusHandler_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodPost, "/api/queue/status", nil)
	rec := httptest.NewRecorder()
	s.queueStatusHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestDocsHandler_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodPost, "/api/docs", nil)
	rec := httptest.NewRecorder()
	s.docsHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestOpenAPIHandler_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodPost, "/api/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	s.openAPIHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestWebhookHandler_DefaultDate(t *testing.T) {
	s := newTestServer(t, "token")
	mux := newTestMux(s)

	req := httptest.NewRequest(http.MethodPost, "/api/load", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	var payload struct {
		Date string `json:"date"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Date == "" {
		t.Fatalf("expected date to be set")
	}
	if _, err := time.Parse("2006-01-02", payload.Date); err != nil {
		t.Fatalf("expected YYYY-MM-DD date, got %s", payload.Date)
	}
}
