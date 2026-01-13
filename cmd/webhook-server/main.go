package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	scalargo "github.com/bdpiprava/scalar-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/user/go-frontol-loader/pkg/auth"
	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/ftp"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/models"
	"github.com/user/go-frontol-loader/pkg/pipeline"
	"github.com/user/go-frontol-loader/pkg/queue"
	"github.com/user/go-frontol-loader/pkg/repository"
	"github.com/user/go-frontol-loader/pkg/validation"
)

// openAPISpec загружает OpenAPI спецификацию
// Используем чтение файла во время выполнения, так как embed не поддерживает пути с ..
var openAPISpec = func() []byte {
	data, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		// Если файл не найден, возвращаем пустой массив
		// В этом случае документация не будет работать, но сервер запустится
		return []byte{}
	}
	return data
}()

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

// TransactionTypeStats представляет статистику по типу транзакций
type TransactionTypeStats struct {
	TableName string `json:"table_name"`
	Count     int    `json:"count"`
}

// WebhookReport представляет отчет о выполнении ETL
type WebhookReport struct {
	RequestID          string                 `json:"request_id"`
	Date               string                 `json:"date"`
	Status             string                 `json:"status"`
	StartTime          time.Time              `json:"start_time"`
	EndTime            time.Time              `json:"end_time"`
	Duration           string                 `json:"duration"`
	FilesProcessed     int                    `json:"files_processed"`
	FilesSkipped       int                    `json:"files_skipped"`
	TransactionsLoaded int                    `json:"transactions_loaded"`
	Errors             int                    `json:"errors"`
	Success            bool                   `json:"success"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	TransactionDetails []TransactionTypeStats `json:"transaction_details,omitempty"` // Детальная информация по типам транзакций
}

// OperationType представляет тип операции
type OperationType string

const (
	OperationTypeLoad     OperationType = "load"     // Загрузка данных из FTP в БД
	OperationTypeDownload OperationType = "download" // Выгрузка данных из БД в файл
)

// QueueItem представляет элемент очереди запросов
type QueueItem struct {
	RequestID     string
	Date          string
	OperationType OperationType // Тип операции: load или download
	SourceFolder  string        // Для операций download - папка кассы
	Logger        *logger.Logger
	CreatedAt     time.Time
	// Для download операций также нужен ResponseWriter и Request для отправки ответа
	DownloadWriter  http.ResponseWriter
	DownloadRequest *http.Request
}

// OperationQueue представляет очередь запросов для конкретного типа операции
// Запросы одного типа операции обрабатываются последовательно (независимо от даты)
type OperationQueue struct {
	operationType OperationType
	queue         chan *QueueItem
	isActive      bool // Флаг, что воркер для этого типа операции активен
	stopChan      chan struct{}
}

// NewOperationQueue создает новую очередь для конкретного типа операции
func NewOperationQueue(operationType OperationType, size int) *OperationQueue {
	if size <= 0 {
		size = 100 // Размер по умолчанию
	}
	return &OperationQueue{
		operationType: operationType,
		queue:         make(chan *QueueItem, size),
		isActive:      false,
		stopChan:      make(chan struct{}),
	}
}

// Enqueue добавляет запрос в очередь
func (oq *OperationQueue) Enqueue(item *QueueItem) error {
	select {
	case oq.queue <- item:
		return nil
	default:
		return fmt.Errorf("queue is full for operation type %s", oq.operationType)
	}
}

// Dequeue извлекает запрос из очереди (блокирующий вызов)
func (oq *OperationQueue) Dequeue() *QueueItem {
	return <-oq.queue
}

// Size возвращает текущий размер очереди
func (oq *OperationQueue) Size() int {
	return len(oq.queue)
}

// Stop останавливает очередь
func (oq *OperationQueue) Stop() {
	close(oq.stopChan)
}

// RequestQueueManager управляет очередями по типам операций
// Разные типы операций обрабатываются параллельно, одинаковые - последовательно
type RequestQueueManager struct {
	operationQueues map[OperationType]*OperationQueue
	mu              sync.RWMutex
	queueSize       int
}

// Server представляет веб-сервер
type Server struct {
	config       *models.Config
	logger       *logger.Logger
	queueManager *RequestQueueManager
	workerStop   chan struct{}
	workerWg     sync.WaitGroup

	// RabbitMQ integration
	queueProvider string
	amqpClient    *queue.Client
	consumerCtx   context.Context
	consumerStop  context.CancelFunc
}

// NewRequestQueueManager создает новый менеджер очередей
func NewRequestQueueManager(queueSize int) *RequestQueueManager {
	return &RequestQueueManager{
		operationQueues: make(map[OperationType]*OperationQueue),
		queueSize:       queueSize,
	}
}

// GetOrCreateQueue получает или создает очередь для указанного типа операции
func (rqm *RequestQueueManager) GetOrCreateQueue(operationType OperationType) *OperationQueue {
	rqm.mu.Lock()
	defer rqm.mu.Unlock()

	if queue, exists := rqm.operationQueues[operationType]; exists {
		return queue
	}

	queue := NewOperationQueue(operationType, rqm.queueSize)
	rqm.operationQueues[operationType] = queue
	return queue
}

// Enqueue добавляет запрос в очередь для соответствующего типа операции
func (rqm *RequestQueueManager) Enqueue(item *QueueItem) error {
	queue := rqm.GetOrCreateQueue(item.OperationType)
	return queue.Enqueue(item)
}

// StartWorkerForOperation запускает воркер для обработки запросов конкретного типа операции
func (rqm *RequestQueueManager) StartWorkerForOperation(operationType OperationType, server *Server) {
	rqm.mu.Lock()
	queue, exists := rqm.operationQueues[operationType]
	if !exists {
		rqm.mu.Unlock()
		return
	}

	// Проверяем, не запущен ли уже воркер для этого типа операции
	if queue.isActive {
		rqm.mu.Unlock()
		return
	}
	queue.isActive = true
	rqm.mu.Unlock()

	// Запускаем воркер в отдельной горутине
	server.workerWg.Add(1)
	go func() {
		defer func() {
			server.workerWg.Done()
			rqm.mu.Lock()
			queue.isActive = false
			rqm.mu.Unlock()
		}()

		server.logger.Info("Operation queue worker started",
			"operation_type", operationType,
			"event", "operation_queue_worker_started",
		)

		for {
			select {
			case <-queue.stopChan:
				server.logger.Info("Operation queue worker stopped",
					"operation_type", operationType,
					"event", "operation_queue_worker_stopped",
				)
				return
			case item := <-queue.queue:
				// Обрабатываем запрос последовательно для этого типа операции
				server.logger.Info("Taking next request from operation queue",
					"request_id", item.RequestID,
					"operation_type", operationType,
					"date", item.Date,
					"queue_size", queue.Size(),
					"event", "operation_queue_item_taken",
				)
				server.processQueueItem(item)

				// После обработки проверяем, есть ли еще запросы в очереди
				// Если очередь пуста, воркер завершится
				queueSize := queue.Size()
				if queueSize == 0 {
					server.logger.Info("Operation queue is empty, worker will stop",
						"operation_type", operationType,
						"event", "operation_queue_empty",
					)
					rqm.mu.Lock()
					queue.isActive = false
					rqm.mu.Unlock()
					return
				}
				// Если есть еще запросы, продолжаем обработку
				server.logger.Info("More requests in queue, continuing processing",
					"operation_type", operationType,
					"remaining_queue_size", queueSize,
					"event", "operation_queue_continuing",
				)
			}
		}
	}()
}

// GetQueueSize возвращает размер очереди для типа операции
func (rqm *RequestQueueManager) GetQueueSize(operationType OperationType) int {
	rqm.mu.RLock()
	defer rqm.mu.RUnlock()

	if queue, exists := rqm.operationQueues[operationType]; exists {
		return queue.Size()
	}
	return 0
}

// GetTotalSize возвращает общий размер всех очередей
func (rqm *RequestQueueManager) GetTotalSize() int {
	rqm.mu.RLock()
	defer rqm.mu.RUnlock()

	total := 0
	for _, queue := range rqm.operationQueues {
		total += queue.Size()
	}
	return total
}

// GetActiveOperationsCount возвращает количество активных типов операций
func (rqm *RequestQueueManager) GetActiveOperationsCount() int {
	rqm.mu.RLock()
	defer rqm.mu.RUnlock()

	return len(rqm.operationQueues)
}

// StopAll останавливает все очереди
func (rqm *RequestQueueManager) StopAll() {
	rqm.mu.Lock()
	defer rqm.mu.Unlock()

	for _, queue := range rqm.operationQueues {
		queue.Stop()
	}
}

// enqueue routes to the configured queue provider (rabbitmq or in-memory).
func (s *Server) enqueue(ctx context.Context, item *QueueItem) error {
	// Download endpoint streams response via in-process worker; keep in-memory queue until async consumer is added.
	if item.OperationType == OperationTypeDownload {
		if err := s.queueManager.Enqueue(item); err != nil {
			return err
		}
		s.queueManager.StartWorkerForOperation(item.OperationType, s)
		return nil
	}

	if s.queueProvider != "rabbitmq" || s.amqpClient == nil {
		// In-memory legacy behavior
		if err := s.queueManager.Enqueue(item); err != nil {
			return err
		}
		s.queueManager.StartWorkerForOperation(item.OperationType, s)
		return nil
	}

	// RabbitMQ publish path
	cashbox := item.SourceFolder
	if cashbox == "" {
		cashbox = "default"
	}

	qs := queue.BuildQueueSet(string(item.OperationType), cashbox)
	backoff := time.Minute
	if len(s.config.QueueRetryBackoffs) > 0 {
		backoff = s.config.QueueRetryBackoffs[0]
	}

	if err := s.amqpClient.DeclareQueues(qs, backoff, s.config.QueueDeclareOnPublish); err != nil {
		return fmt.Errorf("declare topology: %w", err)
	}

	msg := map[string]string{
		"request_id":     item.RequestID,
		"date":           item.Date,
		"operation_type": string(item.OperationType),
		"source_folder":  item.SourceFolder,
		"created_at":     item.CreatedAt.UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	headers := amqp.Table{
		"x-retry-count": int32(0),
		"x-first-seen":  time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.amqpClient.Publish(ctx, qs.RoutingKey, body, headers); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}

func (s *Server) startRabbitConsumers() error {
	// Derive AMQP URL
	amqpURL := s.config.RabbitMQURL
	if amqpURL == "" {
		amqpURL = fmt.Sprintf("amqp://%s:%s@%s:%d%s", s.config.RabbitMQUser, s.config.RabbitMQPassword, s.config.RabbitMQHost, s.config.RabbitMQPort, s.config.RabbitMQVHost)
	}

	consumer := queue.NewConsumer(queue.ConsumerConfig{
		URL:              amqpURL,
		Prefetch:         s.config.RabbitMQPrefetch,
		Backoffs:         s.config.QueueRetryBackoffs,
		MaxRetries:       s.config.QueueRetryMax,
		DeclareOnPublish: s.config.QueueDeclareOnPublish,
	})

	cashboxes := s.flattenCashboxes()
	if len(cashboxes) == 0 {
		cashboxes = []string{"default"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.consumerCtx = ctx
	s.consumerStop = cancel

	go func() {
		if err := consumer.Start(ctx, cashboxes, func(c context.Context, msg queue.Message) error {
			log := s.logger.WithRequestID(msg.RequestID)
			queueItem := &QueueItem{
				RequestID:     msg.RequestID,
				Date:          msg.Date,
				OperationType: OperationType(msg.OperationType),
				SourceFolder:  msg.SourceFolder,
				Logger:        log,
				CreatedAt:     time.Now(),
			}
			if queueItem.OperationType == "" {
				queueItem.OperationType = OperationTypeLoad
			}
			s.runETLPipeline(queueItem.RequestID, queueItem.Date, queueItem.Logger)
			return nil
		}); err != nil {
			s.logger.Error("RabbitMQ consumer stopped with error",
				"error", err.Error(),
				"event", "rabbitmq_consumer_stopped",
			)
		}
	}()

	return nil
}

func (s *Server) flattenCashboxes() []string {
	var cashboxes []string
	for _, folders := range s.config.KassaStructure {
		cashboxes = append(cashboxes, folders...)
	}
	return cashboxes
}

func (s *Server) fetchMgmtQueueStats() (map[string]int, error) {
	client := queue.ManagementClient{
		BaseURL:  s.config.RabbitMQManagementURL,
		Username: s.config.RabbitMQUser,
		Password: s.config.RabbitMQPassword,
	}
	queues, err := client.ListQueues("etl.")
	if err != nil {
		return nil, err
	}
	stats := make(map[string]int)
	for _, q := range queues {
		stats[q.Name] = q.Messages
	}
	return stats, nil
}

func (s *Server) rabbitMQQueueStats(ctx context.Context) (map[string]int, error) {
	if s.amqpClient == nil {
		return nil, fmt.Errorf("amqp client not initialized")
	}

	// Try management API first for consolidated metrics.
	if s.config.RabbitMQManagementURL != "" && s.config.RabbitMQUser != "" {
		stats, err := s.fetchMgmtQueueStats()
		if err == nil {
			return stats, nil
		}
		s.logger.Warn("RabbitMQ management API failed, falling back to passive declare",
			"error", err.Error(),
			"event", "rabbitmq_mgmt_fallback",
		)
	}

	ch, err := s.amqpClient.OpenChannel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	stats := make(map[string]int)
	cashboxes := s.flattenCashboxes()
	if len(cashboxes) == 0 {
		cashboxes = []string{"default"}
	}
	ops := []OperationType{OperationTypeLoad}

	for _, op := range ops {
		for _, cashbox := range cashboxes {
			qs := queue.BuildQueueSet(string(op), cashbox)
			info, err := ch.QueueInspect(qs.PrimaryQueue)
			if err != nil {
				if amqpErr, ok := err.(*amqp.Error); ok && amqpErr.Code == amqp.NotFound {
					s.logger.Warn("Queue not found in RabbitMQ stats",
						"queue", qs.PrimaryQueue,
						"event", "rabbitmq_queue_missing",
					)
					continue
				}
				return nil, fmt.Errorf("inspect queue %s: %w", qs.PrimaryQueue, err)
			}
			stats[qs.PrimaryQueue] = info.Messages
		}
	}

	return stats, nil
}

// requeueDLQHandler triggers DLQ requeue manually.
func (s *Server) requeueDLQHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.queueProvider != "rabbitmq" {
		http.Error(w, "RabbitMQ provider is disabled", http.StatusBadRequest)
		return
	}
	if !s.config.QueueDLQRequeueEnabled {
		http.Error(w, "DLQ requeue is disabled by config", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	q := r.URL.Query()
	op := q.Get("operation")
	cashbox := q.Get("cashbox")
	if op == "" {
		op = string(OperationTypeLoad)
	}
	if cashbox == "" {
		cashbox = "default"
	}

	minAge := s.config.QueueDLQRequeueMinAge
	if ageStr := q.Get("min_age_seconds"); ageStr != "" {
		if v, err := strconv.Atoi(ageStr); err == nil {
			minAge = time.Duration(v) * time.Second
		}
	}
	batch := s.config.QueueDLQRequeueBatch
	if batchStr := q.Get("batch"); batchStr != "" {
		if v, err := strconv.Atoi(batchStr); err == nil && v > 0 {
			batch = v
		}
	}

	qs := queue.BuildQueueSet(op, cashbox)
	requeuer := queue.DLQRequeuer{
		Client: s.amqpClient,
	}
	count, err := requeuer.Requeue(ctx, qs, minAge, batch)
	if err != nil {
		s.logger.Error("DLQ requeue failed",
			"queue", qs.DLQ,
			"error", err.Error(),
			"event", "dlq_requeue_failed",
		)
		http.Error(w, fmt.Sprintf("requeue failed: %v", err), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"queue":        qs.DLQ,
		"requeued":     count,
		"min_age":      minAge.String(),
		"batch":        batch,
		"operation":    op,
		"cashbox":      cashbox,
		"queue_target": qs.PrimaryQueue,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg *models.Config) *Server {
	// Create structured logger
	loggerInstance := logger.New(logger.Config{
		Level:   cfg.LogLevel,
		Format:  "text",
		Output:  os.Stdout,
		Backend: cfg.LogBackend,
	})
	slog.SetDefault(loggerInstance.Logger)

	server := &Server{
		config:        cfg,
		logger:        loggerInstance.WithComponent("webhook-server"),
		queueManager:  NewRequestQueueManager(100), // Очередь на 100 запросов для каждой даты
		workerStop:    make(chan struct{}),
		queueProvider: cfg.QueueProvider,
	}

	if cfg.QueueProvider == "rabbitmq" {
		amqpURL := cfg.RabbitMQURL
		if amqpURL == "" {
			amqpURL = fmt.Sprintf("amqp://%s:%s@%s:%d%s", cfg.RabbitMQUser, cfg.RabbitMQPassword, cfg.RabbitMQHost, cfg.RabbitMQPort, cfg.RabbitMQVHost)
		}
		server.amqpClient = queue.NewClient(queue.Config{
			URL:       amqpURL,
			Prefetch:  cfg.RabbitMQPrefetch,
			Reconnect: 5 * time.Second,
		})
	}

	return server
}

// startWorkerForDate запускает воркер для обработки запросов конкретной даты
// Этот метод вызывается автоматически при добавлении запроса в очередь

// processQueueItem обрабатывает элемент из очереди
// Выполняется строго последовательно - следующий запрос не начнется, пока текущий не завершится полностью
func (s *Server) processQueueItem(item *QueueItem) {
	ctx := context.Background()
	log := item.Logger
	processingStartTime := time.Now()

	log.InfoContext(ctx, "=== STARTING REQUEST PROCESSING ===",
		"request_id", item.RequestID,
		"operation_type", item.OperationType,
		"date", item.Date,
		"queue_wait_time", time.Since(item.CreatedAt).String(),
		"queue_size_before", s.queueManager.GetQueueSize(item.OperationType),
		"total_queue_size", s.queueManager.GetTotalSize(),
		"event", "queue_item_processing_start",
	)

	// Обрабатываем в зависимости от типа операции
	switch item.OperationType {
	case OperationTypeLoad:
		// Запускаем ETL pipeline (блокирующий вызов - ждет полного завершения)
		s.runETLPipeline(item.RequestID, item.Date, log)
	case OperationTypeDownload:
		// Выполняем выгрузку данных
		s.processDownload(item)
	default:
		log.ErrorContext(ctx, "Unknown operation type",
			"operation_type", item.OperationType,
			"event", "unknown_operation_type",
		)
	}

	processingDuration := time.Since(processingStartTime)

	log.InfoContext(ctx, "=== REQUEST PROCESSING COMPLETED ===",
		"request_id", item.RequestID,
		"operation_type", item.OperationType,
		"date", item.Date,
		"processing_duration", processingDuration.String(),
		"queue_size_after", s.queueManager.GetQueueSize(item.OperationType),
		"total_queue_size", s.queueManager.GetTotalSize(),
		"event", "queue_item_processing_completed",
	)

	// После завершения этого запроса воркер для этого типа операции автоматически возьмет следующий из очереди
	// (если он есть). Разные типы операций обрабатываются параллельно.
	log.InfoContext(ctx, "Request removed from operation queue, next request for this operation type will be processed",
		"request_id", item.RequestID,
		"operation_type", item.OperationType,
		"event", "queue_item_removed",
	)
}

// processDownload обрабатывает операцию выгрузки данных
func (s *Server) processDownload(item *QueueItem) {
	ctx := context.Background()
	log := item.Logger

	log.InfoContext(ctx, "Processing download operation",
		"request_id", item.RequestID,
		"source_folder", item.SourceFolder,
		"date", item.Date,
		"event", "download_processing_start",
	)

	// Подключаемся к базе данных
	database, err := db.NewPool(s.config)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database",
			"error", err.Error(),
			"event", "db_connection_error",
		)
		http.Error(item.DownloadWriter, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Создаем loader
	loader := repository.NewLoader(database)

	// Получаем ВСЕ транзакции из ВСЕХ таблиц по source_folder
	allTransactions, err := loader.GetAllTransactionsBySourceFolderAndDate(ctx, item.SourceFolder, item.Date)
	if err != nil {
		log.ErrorContext(ctx, "Failed to get all transactions",
			"error", err.Error(),
			"event", "query_error",
		)
		http.Error(item.DownloadWriter, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	if len(allTransactions) == 0 {
		log.InfoContext(ctx, "No transactions found",
			"source_folder", item.SourceFolder,
			"date", item.Date,
			"event", "no_transactions_found",
		)
		http.Error(item.DownloadWriter, "No transactions found for the specified source_folder and date", http.StatusNotFound)
		return
	}

	log.InfoContext(ctx, "All transactions retrieved",
		"count", len(allTransactions),
		"source_folder", item.SourceFolder,
		"event", "all_transactions_retrieved",
	)

	// Формируем имя файла
	filename := fmt.Sprintf("kassa_%s_%s.txt", strings.ReplaceAll(item.SourceFolder, "/", "_"), item.Date)

	// Создаем TXT файл в формате Frontol
	var buf bytes.Buffer

	// Записываем заголовок файла (3 строки)
	buf.WriteString("#\n")
	buf.WriteString("1\n")
	buf.WriteString(fmt.Sprintf("%d\n", len(allTransactions)))

	// Записываем транзакции
	transactionsWritten := 0
	transactionsSkipped := 0
	skippedTransactionIDs := []int64{}

	for _, t := range allTransactions {
		if t.RawLine != "" {
			buf.WriteString(t.RawLine)
			buf.WriteString("\n")
			transactionsWritten++
		} else {
			transactionsSkipped++
			skippedTransactionIDs = append(skippedTransactionIDs, t.TransactionIDUnique)
			log.WarnContext(ctx, "Transaction without raw line, skipping",
				"transaction_id", t.TransactionIDUnique,
				"transaction_type", t.TransactionType,
				"source_folder", t.SourceFolder,
				"transaction_date", t.TransactionDate,
				"event", "transaction_without_raw_line",
			)
			continue
		}
	}

	log.InfoContext(ctx, "Transactions written to file",
		"written", transactionsWritten,
		"skipped", transactionsSkipped,
		"skipped_ids", skippedTransactionIDs,
		"total_retrieved", len(allTransactions),
		"event", "transactions_write_stats",
	)

	// Устанавливаем заголовки для скачивания файла
	item.DownloadWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	item.DownloadWriter.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	item.DownloadWriter.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	// Отправляем данные
	if _, err := io.Copy(item.DownloadWriter, &buf); err != nil {
		log.ErrorContext(ctx, "Failed to send TXT data",
			"error", err.Error(),
			"event", "txt_send_error",
		)
		return
	}

	log.InfoContext(ctx, "TXT file sent successfully",
		"rows", len(allTransactions),
		"event", "txt_sent",
	)
}

// Stop останавливает сервер и все воркеры очередей с graceful shutdown
func (s *Server) Stop() {
	shutdownTimeout := s.config.ShutdownTimeout
	s.logger.Info("Initiating graceful shutdown",
		"shutdown_timeout", shutdownTimeout.String(),
		"event", "server_stopping",
	)

	if s.consumerStop != nil {
		s.consumerStop()
	}

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Канал для сигнализации завершения shutdown
	done := make(chan struct{})

	go func() {
		// Проверяем текущий размер очередей
		totalQueueSize := s.queueManager.GetTotalSize()
		if totalQueueSize > 0 {
			s.logger.Info("Draining queues before shutdown",
				"total_queue_size", totalQueueSize,
				"load_queue_size", s.queueManager.GetQueueSize(OperationTypeLoad),
				"download_queue_size", s.queueManager.GetQueueSize(OperationTypeDownload),
				"event", "queue_draining_start",
			)
		}

		// Останавливаем прием новых запросов
		close(s.workerStop)

		// Останавливаем все очереди (воркеры завершат текущие задачи)
		s.queueManager.StopAll()

		// Ждем завершения всех воркеров
		s.workerWg.Wait()

		// Логируем финальное состояние очередей
		remainingQueueSize := s.queueManager.GetTotalSize()
		if remainingQueueSize > 0 {
			s.logger.Warn("Shutdown completed with items remaining in queue",
				"remaining_queue_size", remainingQueueSize,
				"event", "queue_not_fully_drained",
			)
		} else {
			s.logger.Info("All queues drained successfully",
				"event", "queue_drained",
			)
		}

		close(done)
	}()

	// Ждем завершения shutdown или таймаута
	select {
	case <-done:
		s.logger.Info("Graceful shutdown completed successfully",
			"event", "server_stopped",
		)
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout reached, forcing shutdown",
			"timeout", shutdownTimeout.String(),
			"remaining_queue_size", s.queueManager.GetTotalSize(),
			"event", "shutdown_timeout",
		)
	}
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

	log.InfoContext(ctx, "Received webhook request",
		"event", "webhook_request",
	)

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrorContext(ctx, "Error reading request body",
			"error", err.Error(),
			"event", "request_read_error",
		)
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

	if err := s.enqueue(ctx, queueItem); err != nil {
		log.ErrorContext(ctx, "Failed to enqueue request",
			"error", err.Error(),
			"operation_type", OperationTypeLoad,
			"date", date,
			"queue_size", s.queueManager.GetQueueSize(OperationTypeLoad),
			"event", "queue_enqueue_error",
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
		"operation_type", OperationTypeLoad,
		"date", date,
		"queue_provider", s.queueProvider,
		"queue_size_for_operation", s.queueManager.GetQueueSize(OperationTypeLoad),
		"total_queue_size", s.queueManager.GetTotalSize(),
		"event", "request_queued",
	)
}

// runETLPipeline запускает ETL pipeline и отправляет отчет
// Отчет отправляется ОДИН РАЗ в любом из случаев:
//   - При успешном завершении pipeline
//   - При ошибке выполнения pipeline
//   - По таймауту (если настроен WEBHOOK_TIMEOUT_MINUTES > 0)
//
// Что произойдет раньше - то и отправит уведомление
func (s *Server) runETLPipeline(requestID, date string, log *logger.Logger) {
	ctx := context.Background()
	startTime := time.Now()
	report := &WebhookReport{
		RequestID: requestID,
		Date:      date,
		StartTime: startTime,
		Status:    "processing",
	}

	log.InfoContext(ctx, "Starting ETL pipeline",
		"date", date,
		"event", "etl_pipeline_start",
	)

	// Каналы для синхронизации
	pipelineDone := make(chan bool, 1)          // Сигнал о завершении pipeline
	reportReady := make(chan *WebhookReport, 1) // Канал для готового отчета
	var reportMutex sync.Mutex                  // Мьютекс для защиты report
	var reportSent bool                         // Флаг, что отчет уже отправлен

	// Pipeline сам управляет таймаутами для загрузки данных через отдельный контекст
	// Это предотвращает отмену контекста при таймауте webhook

	// Запускаем pipeline в горутине
	go func() {
		defer func() {
			pipelineDone <- true
		}()

		result, err := pipeline.Run(ctx, log.Logger, s.config, date)

		reportMutex.Lock()
		if err != nil {
			report.Status = "failed"
			report.ErrorMessage = err.Error()
			report.Success = false
			log.ErrorContext(ctx, "ETL pipeline failed",
				"error", err.Error(),
				"event", "etl_pipeline_failed",
			)
		} else {
			report.Status = "completed"
			report.Success = true
			report.FilesProcessed = result.FilesProcessed
			report.FilesSkipped = result.FilesSkipped
			report.TransactionsLoaded = result.TransactionsLoaded
			report.Errors = result.Errors
			// Преобразуем детальную статистику из pipeline в формат webhook
			report.TransactionDetails = make([]TransactionTypeStats, 0, len(result.TransactionDetails))
			for _, detail := range result.TransactionDetails {
				report.TransactionDetails = append(report.TransactionDetails, TransactionTypeStats{
					TableName: detail.TableName,
					Count:     detail.Count,
				})
			}
			log.InfoContext(ctx, "ETL pipeline completed",
				"files_processed", result.FilesProcessed,
				"transactions_loaded", result.TransactionsLoaded,
				"errors", result.Errors,
				"event", "etl_pipeline_completed",
			)
		}
		reportMutex.Unlock()

		// Отправляем сигнал, что отчет готов
		reportReady <- report
	}()

	// Функция для отправки отчета (только один раз)
	sendReport := func(r *WebhookReport) {
		reportMutex.Lock()
		defer reportMutex.Unlock()
		if !reportSent {
			reportSent = true
			r.EndTime = time.Now()
			r.Duration = r.EndTime.Sub(r.StartTime).String()
			if s.config.WebhookReportURL != "" {
				log.InfoContext(ctx, "Sending webhook report",
					"request_id", requestID,
					"date", date,
					"status", r.Status,
					"event", "webhook_report_sending",
				)
				s.sendWebhookReport(r)
				log.InfoContext(ctx, "Webhook report sent",
					"request_id", requestID,
					"date", date,
					"status", r.Status,
					"event", "webhook_report_sent",
				)
			} else {
				log.InfoContext(ctx, "Webhook report URL not configured, skipping report",
					"request_id", requestID,
					"date", date,
					"event", "webhook_report_skipped",
				)
			}
		}
	}

	// Отправляем отчет либо при завершении pipeline (успешном или с ошибкой), либо по таймауту
	// Отправка происходит только один раз - что произойдет раньше
	if s.config.WebhookTimeoutMinutes == 0 {
		// Если таймаут не настроен (0), отправляем отчет только после завершения pipeline
		// Ждем завершения pipeline (успешного или с ошибкой)
		log.InfoContext(ctx, "Waiting for pipeline completion (no timeout configured)",
			"request_id", requestID,
			"date", date,
			"event", "waiting_pipeline_completion",
		)
		<-pipelineDone
		log.InfoContext(ctx, "Pipeline execution finished, waiting for report",
			"request_id", requestID,
			"date", date,
			"event", "pipeline_done_waiting_report",
		)
		// Ждем готовности отчета
		select {
		case r := <-reportReady:
			sendReport(r)
		case <-time.After(5 * time.Second):
			log.WarnContext(ctx, "Timeout waiting for report",
				"request_id", requestID,
				"date", date,
				"event", "report_timeout",
			)
		}
	} else {
		// Если таймаут настроен, отправляем отчет либо при завершении, либо по таймауту
		// Что произойдет раньше - то и отправит уведомление (только один раз)
		timeout := time.Duration(s.config.WebhookTimeoutMinutes) * time.Minute
		timeoutChan := time.After(timeout)

		select {
		case <-pipelineDone:
			// Pipeline завершился (успешно или с ошибкой) - ждем готовности отчета и отправляем
			log.InfoContext(ctx, "Pipeline execution completed, waiting for report",
				"request_id", requestID,
				"date", date,
				"event", "pipeline_done_waiting_report",
			)
			select {
			case r := <-reportReady:
				sendReport(r)
			case <-time.After(5 * time.Second):
				log.WarnContext(ctx, "Timeout waiting for report",
					"request_id", requestID,
					"date", date,
					"event", "report_timeout",
				)
			}
		case <-timeoutChan:
			// Таймаут сработал - отправляем текущий статус (даже если pipeline еще работает)
			log.WarnContext(ctx, "Webhook timeout reached, sending current status",
				"request_id", requestID,
				"date", date,
				"timeout_minutes", s.config.WebhookTimeoutMinutes,
				"event", "webhook_timeout",
			)
			reportMutex.Lock()
			timeoutReport := &WebhookReport{
				RequestID:          requestID,
				Date:               date,
				StartTime:          startTime,
				Status:             report.Status,
				Success:            report.Success,
				FilesProcessed:     report.FilesProcessed,
				FilesSkipped:       report.FilesSkipped,
				TransactionsLoaded: report.TransactionsLoaded,
				Errors:             report.Errors,
				ErrorMessage:       report.ErrorMessage,
			}
			// Если pipeline еще не завершился, статус будет "processing"
			if timeoutReport.Status == "processing" {
				timeoutReport.Status = "timeout"
				timeoutReport.Success = false
				timeoutReport.ErrorMessage = "Pipeline execution timeout reached"
			}
			reportMutex.Unlock()
			sendReport(timeoutReport)
		}
	}

	reportMutex.Lock()
	status := report.Status
	duration := time.Since(startTime).String()
	reportMutex.Unlock()

	log.InfoContext(ctx, "ETL pipeline execution finished",
		"request_id", requestID,
		"date", date,
		"status", status,
		"duration", duration,
		"event", "etl_pipeline_execution_finished",
	)

	// Логируем, что pipeline полностью завершен и можно обрабатывать следующий запрос
	log.InfoContext(ctx, "Pipeline fully completed, ready for next request",
		"request_id", requestID,
		"date", date,
		"event", "pipeline_fully_completed",
	)
}

// sendWebhookReport отправляет отчет на указанный webhook URL
func (s *Server) sendWebhookReport(report *WebhookReport) {
	// Сериализуем отчет в JSON
	reportJSON, err := json.Marshal(report)
	if err != nil {
		s.logger.Error("Error marshaling report",
			"error", err.Error(),
			"event", "report_marshal_error",
		)
		return
	}

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", s.config.WebhookReportURL, bytes.NewBuffer(reportJSON))
	if err != nil {
		s.logger.Error("Error creating webhook request",
			"error", err.Error(),
			"event", "webhook_request_error",
		)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Frontol-ETL-Webhook/1.0")

	// Добавляем Bearer токен если он настроен
	if s.config.WebhookBearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.WebhookBearerToken)
	}

	// Отправляем запрос
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Error sending webhook report",
			"error", err.Error(),
			"event", "webhook_send_error",
		)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Warn("Failed to close webhook response body",
				"error", err.Error(),
				"event", "webhook_response_close_error",
			)
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.logger.Info("Webhook report sent successfully",
			"status_code", resp.StatusCode,
			"event", "webhook_report_sent",
		)
	} else {
		s.logger.Warn("Webhook report failed",
			"status_code", resp.StatusCode,
			"event", "webhook_report_failed",
		)
	}
}

// healthHandler обрабатывает запросы к /api/health с проверкой зависимостей
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	startTime := time.Now()

	// Проверяем статус зависимостей
	checks := make(map[string]interface{})
	overallStatus := "healthy"

	// Проверка БД
	dbStatus, dbLatency := s.checkDatabase(ctx)
	checks["database"] = map[string]interface{}{
		"status":     dbStatus,
		"latency_ms": dbLatency.Milliseconds(),
	}
	if dbStatus != "healthy" {
		overallStatus = "degraded"
	}

	// Проверка FTP
	ftpStatus, ftpLatency := s.checkFTP(ctx)
	checks["ftp"] = map[string]interface{}{
		"status":     ftpStatus,
		"latency_ms": ftpLatency.Milliseconds(),
	}
	if ftpStatus != "healthy" {
		overallStatus = "degraded"
	}

	// Статус очередей
	checks["queues"] = map[string]interface{}{
		"load_queue_size":     s.queueManager.GetQueueSize(OperationTypeLoad),
		"download_queue_size": s.queueManager.GetQueueSize(OperationTypeDownload),
		"total_queue_size":    s.queueManager.GetTotalSize(),
		"active_operations":   s.queueManager.GetActiveOperationsCount(),
	}

	response := map[string]interface{}{
		"status":           overallStatus,
		"timestamp":        time.Now().Format(time.RFC3339),
		"service":          "frontol-etl-webhook",
		"checks":           checks,
		"response_time_ms": time.Since(startTime).Milliseconds(),
	}

	// Устанавливаем HTTP статус в зависимости от общего статуса
	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

// checkDatabase проверяет подключение к базе данных
func (s *Server) checkDatabase(ctx context.Context) (string, time.Duration) {
	start := time.Now()

	// Пытаемся подключиться к БД с таймаутом
	database, err := db.NewPool(s.config)
	if err != nil {
		return "unhealthy", time.Since(start)
	}
	defer database.Close()

	return "healthy", time.Since(start)
}

// checkFTP проверяет подключение к FTP серверу
func (s *Server) checkFTP(ctx context.Context) (string, time.Duration) {
	start := time.Now()

	// Пытаемся подключиться к FTP с таймаутом
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

// docsHandler обрабатывает запросы к /api/docs - генерирует и отображает HTML страницу с документацией
func (s *Server) docsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем, что спецификация загружена
	if len(openAPISpec) == 0 {
		s.logger.Error("OpenAPI specification not loaded",
			"event", "openapi_spec_not_loaded",
		)
		http.Error(w, "OpenAPI specification not available", http.StatusInternalServerError)
		return
	}

	// Генерируем HTML документацию используя scalar-go
	// Scalar автоматически подхватит security schemes из OpenAPI спецификации
	html, err := scalargo.NewV2(
		scalargo.WithSpecBytes(openAPISpec), // Используем встроенную OpenAPI спецификацию
		scalargo.WithTheme(scalargo.ThemeDefault),
		scalargo.WithSearchHotKey("k"),
		scalargo.WithDefaultHTTPClient("javascript", "fetch"),
		// Аутентификация настроена в OpenAPI спецификации (bearerAuth)
		// Scalar автоматически использует её из спецификации
	)
	if err != nil {
		s.logger.Error("Failed to generate API documentation",
			"error", err.Error(),
			"event", "docs_generation_error",
		)
		http.Error(w, "Failed to generate documentation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// openAPIHandler обрабатывает запросы к /api/openapi.yaml - возвращает OpenAPI спецификацию
func (s *Server) openAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Читаем YAML файл
	yamlContent, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		s.logger.Error("Failed to read openapi.yaml",
			"error", err.Error(),
			"event", "openapi_file_read_error",
		)
		http.Error(w, "OpenAPI specification not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(yamlContent)
}

// queueStatusHandler обрабатывает запросы на получение статуса очереди
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
		"queue_provider":      s.queueProvider,
		"total_queue_size":    s.queueManager.GetTotalSize(),
		"timestamp":           time.Now().Format(time.RFC3339),
		"load_queue_size":     s.queueManager.GetQueueSize(OperationTypeLoad),
		"download_queue_size": s.queueManager.GetQueueSize(OperationTypeDownload),
		"active_operations":   s.queueManager.GetActiveOperationsCount(),
	}

	if s.queueProvider == "rabbitmq" {
		stats, err := s.rabbitMQQueueStats(ctx)
		if err != nil {
			response["rabbitmq_status"] = fmt.Sprintf("error: %v", err)
		} else {
			response["rabbitmq_status"] = "ok"
			perOp := map[string]int{}
			total := 0
			for q, count := range stats {
				total += count
				parts := strings.Split(q, ".")
				if len(parts) >= 3 {
					perOp[parts[1]] += count
				}
			}
			response["rabbitmq"] = map[string]interface{}{
				"queues":        stats,
				"total":         total,
				"per_operation": perOp,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.ErrorContext(ctx, "Failed to encode queue status response",
			"error", err.Error(),
			"event", "queue_status_encode_error",
		)
		return
	}
}

// listKassasHandler обрабатывает запросы на получение списка доступных source_folder
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

	// Подключаемся к базе данных
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

	// Создаем loader
	loader := repository.NewLoader(database)

	// Получаем список source folders
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

	// Формируем ответ
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
		return
	}
}

// downloadHandler обрабатывает запросы на скачивание данных по кассе и дате
func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	log := s.logger.WithRequestID(r.Header.Get("X-Request-ID"))

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
		http.Error(w, fmt.Sprintf("Invalid date: %v", err), http.StatusBadRequest)
		return
	}

	// Генерируем ID запроса для отслеживания
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	log = log.WithRequestID(requestID)

	log.InfoContext(ctx, "Download request received",
		"source_folder", sourceFolder,
		"date", date,
		"event", "download_request",
	)

	// Добавляем запрос в очередь для типа операции "download"
	queueItem := &QueueItem{
		RequestID:       requestID,
		Date:            date,
		OperationType:   OperationTypeDownload,
		SourceFolder:    sourceFolder,
		Logger:          log,
		CreatedAt:       time.Now(),
		DownloadWriter:  w,
		DownloadRequest: r,
	}

	if err := s.enqueue(ctx, queueItem); err != nil {
		log.ErrorContext(ctx, "Failed to enqueue download request",
			"error", err.Error(),
			"operation_type", OperationTypeDownload,
			"date", date,
			"queue_size", s.queueManager.GetQueueSize(OperationTypeDownload),
			"event", "queue_enqueue_error",
		)
		http.Error(w, "Service unavailable: queue is full", http.StatusServiceUnavailable)
		return
	}

	log.InfoContext(ctx, "Download request added to operation queue",
		"operation_type", OperationTypeDownload,
		"date", date,
		"source_folder", sourceFolder,
		"queue_size_for_operation", s.queueManager.GetQueueSize(OperationTypeDownload),
		"total_queue_size", s.queueManager.GetTotalSize(),
		"event", "download_request_queued",
	)
}

// Run запускает веб-сервер
func (s *Server) Run() error {
	// Создаем middleware для Bearer авторизации
	bearerAuth := auth.BearerAuthMiddleware(s.logger.Logger, s.config.WebhookBearerToken)

	// Lazy connect to RabbitMQ if configured
	if s.queueProvider == "rabbitmq" && s.amqpClient != nil {
		if err := s.amqpClient.Connect(); err != nil {
			s.logger.Error("Failed to connect to RabbitMQ, falling back to in-memory queue",
				"error", err.Error(),
				"event", "rabbitmq_connect_error",
			)
			s.queueProvider = "memory"
		} else {
			s.logger.Info("RabbitMQ connected, using broker-backed queues",
				"prefetch", s.config.RabbitMQPrefetch,
				"event", "rabbitmq_connected",
			)
			// Start consumers for load queues
			if err := s.startRabbitConsumers(); err != nil {
				s.logger.Error("Failed to start RabbitMQ consumers, falling back to in-memory queue",
					"error", err.Error(),
					"event", "rabbitmq_consumer_error",
				)
				s.queueProvider = "memory"
			}
		}
	}

	// Настраиваем маршруты с Bearer авторизацией
	// API эндпоинты
	http.HandleFunc("/api/load", bearerAuth(s.webhookHandler))             // POST - загрузка данных из FTP в БД
	http.HandleFunc("/api/files", bearerAuth(s.downloadHandler))           // GET - выгрузка данных из БД в файл
	http.HandleFunc("/api/queue/status", bearerAuth(s.queueStatusHandler)) // GET - статус очереди
	http.HandleFunc("/api/queue/requeue-dlq", bearerAuth(s.requeueDLQHandler))
	http.HandleFunc("/api/kassas", bearerAuth(s.listKassasHandler)) // GET - список касс
	// Health check (без авторизации)
	http.HandleFunc("/api/health", s.healthHandler)

	// Документация API (без авторизации)
	http.HandleFunc("/api/docs", s.docsHandler)
	http.HandleFunc("/api/openapi.yaml", s.openAPIHandler)

	// Статические файлы (без авторизации)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Получаем порт из конфигурации
	port := s.config.ServerPort
	if port == 0 {
		port = 8080 // Порт по умолчанию
	}

	addr := fmt.Sprintf(":%d", port)

	// Логируем состояние авторизации
	if s.config.WebhookBearerToken == "" {
		s.logger.Warn("Bearer token not configured - authorization is DISABLED",
			"event", "auth_config_missing",
		)
	} else {
		s.logger.Info("Bearer token configured - authorization is ENABLED",
			"token_length", len(s.config.WebhookBearerToken),
			"event", "auth_config_loaded",
		)
	}

	s.logger.Info("Starting webhook server",
		"address", addr,
		"event", "server_start",
	)
	s.logger.Info("Queue system: parallel processing for different operation types, sequential for same operation type",
		"event", "queue_system_info",
	)
	s.logger.Info("Available endpoints",
		"endpoints", []string{
			"POST /api/load - загрузка данных из FTP в БД",
			"GET /api/files?source_folder=XXX&date=YYYY-MM-DD - выгрузка данных из БД в файл",
			"GET /api/queue/status - статус очереди",
			"GET /api/kassas - список касс",
			"GET /api/health - health check",
			"GET /api/docs - документация API (Scalar)",
			"GET /api/openapi.yaml - OpenAPI спецификация",
		},
		"event", "server_endpoints",
	)

	// Запускаем сервер с таймаутами
	server := &http.Server{
		Addr:              addr,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return server.ListenAndServe()
}

func main() {
	defaultLogger := logger.New(logger.Config{
		Level:   "info",
		Format:  "text",
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
