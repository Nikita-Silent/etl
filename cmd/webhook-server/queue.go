package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/models"
)

// OperationType представляет тип операции.
type OperationType string

const (
	OperationTypeLoad     OperationType = "load"
	OperationTypeDownload OperationType = "download"
)

// QueueItem представляет элемент очереди запросов.
type QueueItem struct {
	RequestID     string
	Date          string
	OperationType OperationType
	SourceFolder  string
	Logger        *logger.Logger
	CreatedAt     time.Time
}

// OperationQueue представляет очередь запросов для конкретного типа операции.
type OperationQueue struct {
	operationType OperationType
	queue         chan *QueueItem
	mu            sync.Mutex
	workerStarted bool
	closed        bool
}

func NewOperationQueue(operationType OperationType, size int) *OperationQueue {
	if size <= 0 {
		size = 100
	}
	return &OperationQueue{
		operationType: operationType,
		queue:         make(chan *QueueItem, size),
	}
}

func (oq *OperationQueue) Enqueue(item *QueueItem) error {
	oq.mu.Lock()
	defer oq.mu.Unlock()
	if oq.closed {
		return fmt.Errorf("queue is closed for operation type %s", oq.operationType)
	}
	select {
	case oq.queue <- item:
		return nil
	default:
		return fmt.Errorf("queue is full for operation type %s", oq.operationType)
	}
}

func (oq *OperationQueue) Size() int {
	return len(oq.queue)
}

func (oq *OperationQueue) Stop() {
	oq.mu.Lock()
	defer oq.mu.Unlock()
	if oq.closed {
		return
	}
	oq.closed = true
	close(oq.queue)
}

// RequestQueueManager управляет очередями по типам операций.
type RequestQueueManager struct {
	operationQueues map[OperationType]*OperationQueue
	mu              sync.RWMutex
	queueSize       int
}

// Server представляет веб-сервер.
type Server struct {
	config       *models.Config
	logger       *logger.Logger
	queueManager *RequestQueueManager
	workerWg     sync.WaitGroup
	httpServer   *http.Server
	stopping     atomic.Bool
}

func NewRequestQueueManager(queueSize int) *RequestQueueManager {
	return &RequestQueueManager{
		operationQueues: make(map[OperationType]*OperationQueue),
		queueSize:       queueSize,
	}
}

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

func (rqm *RequestQueueManager) Enqueue(item *QueueItem) error {
	queue := rqm.GetOrCreateQueue(item.OperationType)
	return queue.Enqueue(item)
}

func (rqm *RequestQueueManager) StartWorkerForOperation(operationType OperationType, server *Server) {
	rqm.mu.Lock()
	queue, exists := rqm.operationQueues[operationType]
	if !exists {
		rqm.mu.Unlock()
		return
	}

	if queue.workerStarted {
		rqm.mu.Unlock()
		return
	}
	queue.workerStarted = true
	rqm.mu.Unlock()

	server.workerWg.Add(1)
	go func() {
		defer server.workerWg.Done()

		server.logger.Info("Operation queue worker started",
			"operation_type", operationType,
			"event", "operation_queue_worker_started",
		)

		for item := range queue.queue {
			server.logger.Info("Taking next request from operation queue",
				"request_id", item.RequestID,
				"operation_type", operationType,
				"date", item.Date,
				"queue_size", queue.Size(),
				"event", "operation_queue_item_taken",
			)
			server.processQueueItem(item)
		}

		server.logger.Info("Operation queue worker stopped",
			"operation_type", operationType,
			"event", "operation_queue_worker_stopped",
		)
	}()
}

func (rqm *RequestQueueManager) GetQueueSize(operationType OperationType) int {
	rqm.mu.RLock()
	defer rqm.mu.RUnlock()

	if queue, exists := rqm.operationQueues[operationType]; exists {
		return queue.Size()
	}
	return 0
}

func (rqm *RequestQueueManager) GetTotalSize() int {
	rqm.mu.RLock()
	defer rqm.mu.RUnlock()

	total := 0
	for _, queue := range rqm.operationQueues {
		if queue.workerStarted && !queue.closed {
			total++
		}
	}
	return total
}

func (rqm *RequestQueueManager) GetActiveOperationsCount() int {
	rqm.mu.RLock()
	defer rqm.mu.RUnlock()
	return len(rqm.operationQueues)
}

func (rqm *RequestQueueManager) StopAll() {
	rqm.mu.Lock()
	defer rqm.mu.Unlock()

	for _, queue := range rqm.operationQueues {
		queue.Stop()
	}
}

func (s *Server) enqueue(item *QueueItem) error {
	if s.stopping.Load() {
		return fmt.Errorf("server is shutting down")
	}
	if err := s.queueManager.Enqueue(item); err != nil {
		return err
	}
	s.queueManager.StartWorkerForOperation(item.OperationType, s)
	return nil
}

func NewServer(cfg *models.Config) *Server {
	loggerInstance := logger.New(logger.Config{
		Level:   cfg.LogLevel,
		Format:  cfg.LogFormat,
		Output:  os.Stdout,
		Backend: cfg.LogBackend,
	})
	slog.SetDefault(loggerInstance.Logger)

	return &Server{
		config:       cfg,
		logger:       loggerInstance.WithComponent("webhook-server"),
		queueManager: NewRequestQueueManager(100),
	}
}

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

	switch item.OperationType {
	case OperationTypeLoad:
		s.runETLPipeline(item.RequestID, item.Date, log)
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

	log.InfoContext(ctx, "Request removed from operation queue, next request for this operation type will be processed",
		"request_id", item.RequestID,
		"operation_type", item.OperationType,
		"event", "queue_item_removed",
	)
}

func (s *Server) Stop() {
	if !s.stopping.CompareAndSwap(false, true) {
		return
	}

	shutdownTimeout := s.config.EffectiveShutdownTimeout()
	s.logger.Info("Initiating graceful shutdown",
		"shutdown_timeout", shutdownTimeout.String(),
		"event", "server_stopping",
	)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		if s.httpServer != nil {
			if err := s.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Warn("HTTP server shutdown returned error",
					"error", err.Error(),
					"event", "http_server_shutdown_error",
				)
			}
		}

		totalQueueSize := s.queueManager.GetTotalSize()
		if totalQueueSize > 0 {
			s.logger.Info("Draining queues before shutdown",
				"total_queue_size", totalQueueSize,
				"load_queue_size", s.queueManager.GetQueueSize(OperationTypeLoad),
				"download_queue_size", s.queueManager.GetQueueSize(OperationTypeDownload),
				"event", "queue_draining_start",
			)
		}

		s.queueManager.StopAll()
		s.workerWg.Wait()

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
