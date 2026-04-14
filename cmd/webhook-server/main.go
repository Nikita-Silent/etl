package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/validation"
)

// WebhookRequest представляет запрос к webhook
type WebhookRequest struct {
	Date string `json:"date"`
}

// WebhookResponse представляет ответ webhook
type WebhookResponse struct {
	Status    string `json:"status"`
	Date      string `json:"date"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// webhookHandler обрабатывает POST запросы к /api/load
func (s *Server) webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Генерируем ID запроса для отслеживания
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	ctx := r.Context()
	log := s.logger.WithRequestID(requestID)
	audit := newRequestAudit(requestID, "/api/load", string(OperationTypeLoad), r)

	logAPIRequestReceived(ctx, log, audit)

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrorContext(ctx, "Error reading request body",
			"error", err.Error(),
			"event", "request_read_error",
		)
		logAPIRequestRejected(ctx, log, audit, http.StatusBadRequest, "request_body_read_error")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.WarnContext(ctx, "Failed to close request body",
				"error", err.Error(),
				"event", "request_body_close_error",
			)
		}
	}()

	// Парсим JSON
	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.ErrorContext(ctx, "Error parsing JSON",
			"error", err.Error(),
			"event", "json_parse_error",
		)
		logAPIRequestRejected(ctx, log, audit, http.StatusBadRequest, "invalid_json")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидируем дату используя validation framework
	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	} else {
		// Создаем композитный валидатор для даты
		dateValidator := validation.NewComposite(
			validation.DateFormat("date", "2006-01-02"),
			validation.NotInFuture("date", "2006-01-02"),
		)

		if err := dateValidator.Validate(date); err != nil {
			log.ErrorContext(ctx, "Date validation failed",
				"date", date,
				"error", err.Error(),
				"event", "date_validation_error",
			)
			logAPIRequestRejected(ctx, log, audit, http.StatusBadRequest, "invalid_date", "date", date)
			http.Error(w, fmt.Sprintf("Invalid date: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Добавляем запрос в очередь для типа операции "load"
	queueItem := &QueueItem{
		RequestID:     requestID,
		Date:          date,
		OperationType: OperationTypeLoad,
		Logger:        log,
		CreatedAt:     time.Now(),
	}

	if err := s.enqueue(queueItem); err != nil {
		log.ErrorContext(ctx, "Failed to enqueue request",
			"error", err.Error(),
			"operation_type", OperationTypeLoad,
			"date", date,
			"queue_size", s.queueManager.GetQueueSize(OperationTypeLoad),
			"event", "queue_enqueue_error",
		)
		logAPIRequestRejected(ctx, log, audit, http.StatusServiceUnavailable, "queue_unavailable",
			"date", date,
			"queue_size_for_operation", s.queueManager.GetQueueSize(OperationTypeLoad),
			"total_queue_size", s.queueManager.GetTotalSize(),
		)
		http.Error(w, "Service unavailable: queue is full", http.StatusServiceUnavailable)
		return
	}

	// Отвечаем клиенту немедленно
	response := WebhookResponse{
		Status:    "queued",
		Date:      date,
		Message:   "Request added to queue",
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.ErrorContext(ctx, "Error encoding response",
			"error", err.Error(),
			"event", "response_encode_error",
		)
		return
	}

	log.InfoContext(ctx, "Request added to operation queue",
		"log_kind", "loki_operational",
		"operation_type", OperationTypeLoad,
		"date", date,
		"queue_size_for_operation", s.queueManager.GetQueueSize(OperationTypeLoad),
		"total_queue_size", s.queueManager.GetTotalSize(),
		"event", "request_queued",
	)
	logAPIRequestCompleted(ctx, log, audit, http.StatusAccepted, "queued",
		"date", date,
		"queue_size_for_operation", s.queueManager.GetQueueSize(OperationTypeLoad),
		"total_queue_size", s.queueManager.GetTotalSize(),
	)
}

// downloadHandler обрабатывает запросы на скачивание данных по кассе и дате
func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	log := s.logger.WithRequestID(requestID)
	audit := newRequestAudit(requestID, "/api/files", string(OperationTypeDownload), r)
	logAPIRequestReceived(ctx, log, audit)

	// Получаем параметры из query string
	sourceFolder := r.URL.Query().Get("source_folder")
	date := r.URL.Query().Get("date")

	// Валидация параметров используя validation framework
	// Валидация source_folder
	sourceFolderValidator := validation.NewComposite(
		validation.Required("source_folder"),
		validation.KassaCode("source_folder"),
	)
	if err := sourceFolderValidator.Validate(sourceFolder); err != nil {
		log.ErrorContext(ctx, "source_folder validation failed",
			"source_folder", sourceFolder,
			"error", err.Error(),
			"event", "source_folder_validation_error",
		)
		logAPIRequestRejected(ctx, log, audit, http.StatusBadRequest, "invalid_source_folder", "source_folder", sourceFolder)
		http.Error(w, fmt.Sprintf("Invalid source_folder: %v", err), http.StatusBadRequest)
		return
	}

	// Валидация даты
	dateValidator := validation.NewComposite(
		validation.Required("date"),
		validation.DateFormat("date", "2006-01-02"),
		validation.NotInFuture("date", "2006-01-02"),
	)
	if err := dateValidator.Validate(date); err != nil {
		log.ErrorContext(ctx, "date validation failed",
			"date", date,
			"error", err.Error(),
			"event", "date_validation_error",
		)
		logAPIRequestRejected(ctx, log, audit, http.StatusBadRequest, "invalid_date", "date", date)
		http.Error(w, fmt.Sprintf("Invalid date: %v", err), http.StatusBadRequest)
		return
	}

	log.InfoContext(ctx, "Download request received",
		"log_kind", "loki_operational",
		"source_folder", sourceFolder,
		"date", date,
		"event", "download_request",
	)

	if s.stopping.Load() {
		log.WarnContext(ctx, "Rejecting download request during shutdown",
			"source_folder", sourceFolder,
			"date", date,
			"event", "download_rejected_shutdown",
		)
		logAPIRequestRejected(ctx, log, audit, http.StatusServiceUnavailable, "server_shutting_down",
			"source_folder", sourceFolder,
			"date", date,
		)
		http.Error(w, "Service unavailable: server is shutting down", http.StatusServiceUnavailable)
		return
	}

	log.InfoContext(ctx, "Starting synchronous download request",
		"log_kind", "loki_operational",
		"date", date,
		"source_folder", sourceFolder,
		"event", "download_request_processing",
	)
	result := s.processDownloadRequest(ctx, log, w, sourceFolder, date)
	if result.StatusCode >= 400 {
		logAPIRequestRejected(ctx, log, audit, result.StatusCode, result.Outcome,
			"source_folder", sourceFolder,
			"date", date,
			"rows_retrieved", result.RowsRetrieved,
			"rows_written", result.RowsWritten,
			"rows_skipped", result.RowsSkipped,
		)
		return
	}
	logAPIRequestCompleted(ctx, log, audit, result.StatusCode, result.Outcome,
		"source_folder", sourceFolder,
		"date", date,
		"rows_retrieved", result.RowsRetrieved,
		"rows_written", result.RowsWritten,
		"rows_skipped", result.RowsSkipped,
		"skipped_ids_truncated", result.SkippedIDsTruncated,
	)
}

func main() {
	defaultLogger := logger.New(logger.Config{
		Level:   "info",
		Format:  os.Getenv("LOG_FORMAT"),
		Output:  os.Stdout,
		Backend: os.Getenv("LOG_BACKEND"),
	})
	slog.SetDefault(defaultLogger.Logger)

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	// Создаем и запускаем сервер
	server := NewServer(cfg)
	if err := server.Run(); err != nil {
		slog.Error("Server failed",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
