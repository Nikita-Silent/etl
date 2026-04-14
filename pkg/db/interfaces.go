package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// DatabasePool defines the interface for database operations
// This allows for easier testing with mocks
type DatabasePool interface {
	Close()
	BeginTx(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	// Load methods used by the current schema-driven loader path.
	LoadData(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error
	LoadTxTable(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error
}

// Ensure Pool implements DatabasePool interface
var _ DatabasePool = (*Pool)(nil)
