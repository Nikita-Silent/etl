package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/ftp"
	"github.com/user/go-frontol-loader/pkg/repository"
)

// healthHandler обрабатывает запросы к /api/health с проверкой зависимостей.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	startTime := time.Now()
	checks := make(map[string]interface{})
	overallStatus := "healthy"

	dbStatus, dbLatency := s.checkDatabase(ctx)
	checks["database"] = map[string]interface{}{
		"status":     dbStatus,
		"latency_ms": dbLatency.Milliseconds(),
	}
	if dbStatus != "healthy" {
		overallStatus = "degraded"
	}

	ftpStatus, ftpLatency := s.checkFTP(ctx)
	checks["ftp"] = map[string]interface{}{
		"status":     ftpStatus,
		"latency_ms": ftpLatency.Milliseconds(),
	}
	if ftpStatus != "healthy" {
		overallStatus = "degraded"
	}

	checks["queues"] = map[string]interface{}{
		"load_queue_size":     s.queueManager.GetQueueSize(OperationTypeLoad),
		"download_queue_size": s.queueManager.GetQueueSize(OperationTypeDownload),
		"total_queue_size":    s.queueManager.GetTotalSize(),
		"active_operations":   s.queueManager.GetActiveOperationsCount(),
		"is_shutting_down":    s.stopping.Load(),
	}

	response := map[string]interface{}{
		"status":           overallStatus,
		"timestamp":        time.Now().Format(time.RFC3339),
		"service":          "frontol-etl-webhook",
		"checks":           checks,
		"response_time_ms": time.Since(startTime).Milliseconds(),
	}

	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) checkDatabase(ctx context.Context) (string, time.Duration) {
	start := time.Now()
	database, err := db.NewPool(s.config)
	if err != nil {
		return "unhealthy", time.Since(start)
	}
	defer database.Close()
	return "healthy", time.Since(start)
}

func (s *Server) checkFTP(ctx context.Context) (string, time.Duration) {
	start := time.Now()
	ftpClient, err := ftp.NewClient(s.config)
	if err != nil {
		return "unhealthy", time.Since(start)
	}
	defer func() {
		if err := ftpClient.Close(); err != nil {
			s.logger.Warn("Failed to close FTP client",
				"error", err.Error(),
				"event", "ftp_client_close_error",
			)
		}
	}()
	return "healthy", time.Since(start)
}

// queueStatusHandler обрабатывает запросы на получение статуса очереди.
func (s *Server) queueStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	log := s.logger.WithRequestID(r.Header.Get("X-Request-ID"))

	log.InfoContext(ctx, "Queue status request received",
		"event", "queue_status_request",
	)

	response := map[string]interface{}{
		"queue_provider":      "memory",
		"total_queue_size":    s.queueManager.GetTotalSize(),
		"timestamp":           time.Now().Format(time.RFC3339),
		"load_queue_size":     s.queueManager.GetQueueSize(OperationTypeLoad),
		"download_queue_size": s.queueManager.GetQueueSize(OperationTypeDownload),
		"active_operations":   s.queueManager.GetActiveOperationsCount(),
		"is_shutting_down":    s.stopping.Load(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.ErrorContext(ctx, "Failed to encode queue status response",
			"error", err.Error(),
			"event", "queue_status_encode_error",
		)
	}
}

// listKassasHandler обрабатывает запросы на получение списка доступных source_folder.
func (s *Server) listKassasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	log := s.logger.WithRequestID(r.Header.Get("X-Request-ID"))

	log.InfoContext(ctx, "List source folders request received",
		"event", "list_source_folders_request",
	)

	database, err := db.NewPool(s.config)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database",
			"error", err.Error(),
			"event", "db_connection_error",
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	loader := repository.NewLoader(database)
	sourceFolders, err := loader.GetAvailableSourceFolders(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to get source folders",
			"error", err.Error(),
			"event", "query_error",
		)
		http.Error(w, "Failed to retrieve source folders", http.StatusInternalServerError)
		return
	}

	log.InfoContext(ctx, "Source folders retrieved",
		"count", len(sourceFolders),
		"event", "source_folders_retrieved",
	)

	response := map[string]interface{}{
		"source_folders": sourceFolders,
		"count":          len(sourceFolders),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.ErrorContext(ctx, "Failed to encode response",
			"error", err.Error(),
			"event", "response_encode_error",
		)
	}
}
