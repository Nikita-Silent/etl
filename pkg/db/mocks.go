package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// MockPool is a mock implementation of DatabasePool for testing
type MockPool struct {
	BeginTxFunc     func(ctx context.Context) (pgx.Tx, error)
	QueryFunc       func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRowFunc    func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	LoadDataFunc    func(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error
	LoadTxTableFunc func(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error
}

func (m *MockPool) Close() {}

func (m *MockPool) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return nil, nil
}

func (m *MockPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

func (m *MockPool) LoadData(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error {
	if m.LoadDataFunc != nil {
		return m.LoadDataFunc(ctx, tx, tableName, columns, rows)
	}
	return nil
}

func (m *MockPool) LoadTxTable(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
	if m.LoadTxTableFunc != nil {
		return m.LoadTxTableFunc(ctx, tx, tableName, data)
	}
	return nil
}
