package operations

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/models"
)

type Status string

const (
	StatusStarted         Status = "started"
	StatusQueued          Status = "queued"
	StatusProcessing      Status = "processing"
	StatusCompleted       Status = "completed"
	StatusPartial         Status = "partial"
	StatusFailed          Status = "failed"
	StatusTimeoutReported Status = "timeout_reported"
	StatusAbandoned       Status = "abandoned"
)

type executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Close()
}

type opener func(cfg *models.Config) (executor, error)

type Store struct {
	cfg        *models.Config
	log        *logger.Logger
	instanceID string

	once    sync.Once
	pool    executor
	poolErr error
	openFn  opener
}

type Record struct {
	OperationID       string
	RequestID         string
	OperationType     string
	Status            Status
	Date              string
	SourceFolder      string
	Component         string
	StartedAt         time.Time
	UpdatedAt         time.Time
	FinishedAt        *time.Time
	ErrorMessage      string
	FailedStage       string
	TimeoutReportSent bool
	CrashSuspected    bool
}

func NewStore(cfg *models.Config, log *logger.Logger) *Store {
	instanceID, err := os.Hostname()
	if err != nil || instanceID == "" {
		instanceID = "unknown"
	}
	return &Store{
		cfg:        cfg,
		log:        log.WithComponent("operation-store"),
		instanceID: instanceID,
		openFn: func(cfg *models.Config) (executor, error) {
			return db.NewPool(cfg)
		},
	}
}

func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

func (s *Store) Start(ctx context.Context, record Record) error {
	if record.StartedAt.IsZero() {
		record.StartedAt = time.Now()
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = record.StartedAt
	}
	if record.Status == "" {
		record.Status = StatusStarted
	}
	return s.upsert(ctx, record)
}

func (s *Store) Update(ctx context.Context, record Record) error {
	if record.OperationID == "" {
		return fmt.Errorf("operation_id is required")
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = time.Now()
	}
	return s.upsert(ctx, record)
}

func (s *Store) RecoverStale(ctx context.Context) (int, error) {
	pool, err := s.poolOrErr()
	if err != nil {
		return 0, err
	}
	cutoff := time.Now().Add(-s.cfg.EffectiveOperationStaleTimeout())
	rows, err := pool.Query(ctx, `
		UPDATE etl_operation_runs
		SET status = $1,
		    crash_suspected = TRUE,
		    updated_at = NOW(),
		    finished_at = COALESCE(finished_at, NOW()),
		    error_message = COALESCE(NULLIF(error_message, ''), $2),
		    failed_stage = COALESCE(NULLIF(failed_stage, ''), $3)
		WHERE status = ANY($4)
		  AND updated_at < $5
		RETURNING operation_id
	`, StatusAbandoned, "operation abandoned after restart or stale timeout", "stale_operation_recovery", []string{string(StatusStarted), string(StatusQueued), string(StatusProcessing), string(StatusTimeoutReported)}, cutoff)
	if err != nil {
		return 0, fmt.Errorf("recover stale operations: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if rows.Err() != nil {
		return count, fmt.Errorf("iterate stale operations: %w", rows.Err())
	}
	return count, nil
}

func (s *Store) poolOrErr() (executor, error) {
	s.once.Do(func() {
		s.pool, s.poolErr = s.openFn(s.cfg)
		if s.poolErr != nil {
			s.log.Warn("Operation registry disabled: failed to open database connection",
				"error", s.poolErr.Error(),
				"event", "operation_registry_disabled",
			)
		}
	})
	if s.poolErr != nil {
		return nil, s.poolErr
	}
	return s.pool, nil
}

func (s *Store) upsert(ctx context.Context, record Record) error {
	pool, err := s.poolOrErr()
	if err != nil {
		return err
	}
	if record.Component == "" {
		record.Component = "etl"
	}
	if record.StartedAt.IsZero() {
		record.StartedAt = record.UpdatedAt
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO etl_operation_runs (
			operation_id,
			request_id,
			operation_type,
			status,
			date,
			source_folder,
			component,
			instance_id,
			started_at,
			updated_at,
			finished_at,
			error_message,
			failed_stage,
			timeout_report_sent,
			crash_suspected
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (operation_id) DO UPDATE SET
			request_id = COALESCE(NULLIF(EXCLUDED.request_id, ''), etl_operation_runs.request_id),
			operation_type = COALESCE(NULLIF(EXCLUDED.operation_type, ''), etl_operation_runs.operation_type),
			status = EXCLUDED.status,
			date = COALESCE(NULLIF(EXCLUDED.date, ''), etl_operation_runs.date),
			source_folder = COALESCE(NULLIF(EXCLUDED.source_folder, ''), etl_operation_runs.source_folder),
			component = COALESCE(NULLIF(EXCLUDED.component, ''), etl_operation_runs.component),
			instance_id = COALESCE(NULLIF(EXCLUDED.instance_id, ''), etl_operation_runs.instance_id),
			updated_at = EXCLUDED.updated_at,
			finished_at = COALESCE(EXCLUDED.finished_at, etl_operation_runs.finished_at),
			error_message = COALESCE(NULLIF(EXCLUDED.error_message, ''), etl_operation_runs.error_message),
			failed_stage = COALESCE(NULLIF(EXCLUDED.failed_stage, ''), etl_operation_runs.failed_stage),
			timeout_report_sent = etl_operation_runs.timeout_report_sent OR EXCLUDED.timeout_report_sent,
			crash_suspected = EXCLUDED.crash_suspected
	`,
		record.OperationID,
		record.RequestID,
		record.OperationType,
		record.Status,
		record.Date,
		record.SourceFolder,
		record.Component,
		s.instanceID,
		record.StartedAt,
		record.UpdatedAt,
		record.FinishedAt,
		record.ErrorMessage,
		record.FailedStage,
		record.TimeoutReportSent,
		record.CrashSuspected,
	)
	if err != nil {
		return fmt.Errorf("upsert operation run: %w", err)
	}
	return nil
}
