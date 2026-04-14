package operations

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/models"
)

type stubExecutor struct {
	execCalls  int
	queryCalls int
	execErr    error
	queryErr   error
	rows       pgx.Rows
}

func (s *stubExecutor) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	s.execCalls++
	return pgconn.CommandTag{}, s.execErr
}

func (s *stubExecutor) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	s.queryCalls++
	return s.rows, s.queryErr
}

func (s *stubExecutor) Close() {}

type stubRows struct {
	remaining int
}

func (s *stubRows) Close()                                       {}
func (s *stubRows) Err() error                                   { return nil }
func (s *stubRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (s *stubRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (s *stubRows) Next() bool {
	if s.remaining == 0 {
		return false
	}
	s.remaining--
	return true
}
func (s *stubRows) Scan(dest ...any) error { return nil }
func (s *stubRows) Values() ([]any, error) { return nil, nil }
func (s *stubRows) RawValues() [][]byte    { return nil }
func (s *stubRows) Conn() *pgx.Conn        { return nil }

func newTestStore(t *testing.T, exec executor) *Store {
	t.Helper()
	buf := &bytes.Buffer{}
	store := NewStore(&models.Config{OperationStaleTimeout: 2 * time.Hour}, logger.New(logger.Config{Output: buf, Format: "json"}))
	store.openFn = func(cfg *models.Config) (executor, error) {
		return exec, nil
	}
	return store
}

func TestStoreStartUpsertsRecord(t *testing.T) {
	exec := &stubExecutor{}
	store := newTestStore(t, exec)
	if err := store.Start(context.Background(), Record{OperationID: "op_1", OperationType: "load", Status: StatusStarted, Component: "webhook-server"}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if exec.execCalls != 1 {
		t.Fatalf("execCalls = %d, want 1", exec.execCalls)
	}
}

func TestStoreRecoverStaleCountsRows(t *testing.T) {
	exec := &stubExecutor{rows: &stubRows{remaining: 2}}
	store := newTestStore(t, exec)
	count, err := store.RecoverStale(context.Background())
	if err != nil {
		t.Fatalf("RecoverStale() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("RecoverStale() count = %d, want 2", count)
	}
	if exec.queryCalls != 1 {
		t.Fatalf("queryCalls = %d, want 1", exec.queryCalls)
	}
}

func TestStoreUpdateReturnsOpenError(t *testing.T) {
	buf := &bytes.Buffer{}
	store := NewStore(&models.Config{OperationStaleTimeout: 2 * time.Hour}, logger.New(logger.Config{Output: buf, Format: "json"}))
	store.openFn = func(cfg *models.Config) (executor, error) {
		return nil, errors.New("boom")
	}
	if err := store.Update(context.Background(), Record{OperationID: "op_1", Status: StatusFailed}); err == nil {
		t.Fatal("Update() expected error, got nil")
	}
}
