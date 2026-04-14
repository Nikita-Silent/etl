package db

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/go-frontol-loader/pkg/models"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// convertToUTF8 converts a string from Windows-1251 to UTF-8
// If conversion fails, returns the original string
func convertToUTF8(s string) string {
	if s == "" {
		return s
	}

	// Try to convert from Windows-1251 to UTF-8
	decoder := charmap.Windows1251.NewDecoder()
	result, _, err := transform.String(decoder, s)
	if err != nil {
		// If conversion fails, try to clean invalid UTF-8 bytes
		// This handles cases where the string is already UTF-8 or mixed encoding
		return strings.ToValidUTF8(s, "")
	}
	return result
}

// safeValue returns a safe value for database insertion, converting empty strings to nil
// Also converts string values from Windows-1251 to UTF-8
func safeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		// Convert from Windows-1251 to UTF-8 before inserting into database
		return convertToUTF8(v)
	case float64:
		if v == 0.0 {
			return nil
		}
		return v
	case int:
		if v == 0 {
			return nil
		}
		return v
	case int64:
		if v == 0 {
			return nil
		}
		return v
	default:
		return value
	}
}

// safeValueAllowZero behaves like safeValue, but preserves numeric zero values.
func safeValueAllowZero(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		return convertToUTF8(v)
	case float64:
		return v
	case int:
		return v
	case int64:
		return v
	default:
		return value
	}
}

// Pool represents the database connection pool
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a new database connection pool
func NewPool(cfg *models.Config) (*Pool, error) {
	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	// Configure connection pool
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}
	config.ConnConfig.ConnectTimeout = cfg.EffectiveDBConnectTimeout()

	// Set pool configuration
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	// Create pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.EffectiveDBConnectTimeout())
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{pool}, nil
}

// Close closes the connection pool
func (p *Pool) Close() {
	p.Pool.Close()
}

// BeginTx starts a new transaction with explicit isolation level
// Uses READ COMMITTED isolation level (default for PostgreSQL, suitable for ETL operations)
func (p *Pool) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return p.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
}

// Query executes a query and returns rows
func (p *Pool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.Pool.Query(ctx, sql, args...)
}

// QueryRow executes a query and returns a single row
func (p *Pool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.Pool.QueryRow(ctx, sql, args...)
}

// Exec executes a query without returning rows.
func (p *Pool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, sql, args...)
}

// convertRowValues converts all string values in a row from Windows-1251 to UTF-8
func convertRowValues(row []interface{}) []interface{} {
	converted := make([]interface{}, len(row))
	for i, val := range row {
		if str, ok := val.(string); ok {
			converted[i] = convertToUTF8(str)
		} else {
			converted[i] = val
		}
	}
	return converted
}

// LoadData loads data into the database using INSERT ... ON CONFLICT
// Uses transaction for atomicity - all operations in a transaction are committed or rolled back together
func (p *Pool) LoadData(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// Build placeholders for the query
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Build column list with quotes
	columnList := make([]string, len(columns))
	for i, col := range columns {
		columnList[i] = fmt.Sprintf(`"%s"`, col)
	}

	// Build UPDATE clause for ON CONFLICT
	updateClause := make([]string, 0)
	for _, col := range columns {
		if col != "transaction_id_unique" && col != "source_folder" {
			updateClause = append(updateClause, fmt.Sprintf(`"%s" = EXCLUDED."%s"`, col, col))
		}
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) 
		VALUES (%s) 
		ON CONFLICT (transaction_id_unique, source_folder) 
		DO UPDATE SET %s
	`, tableName,
		strings.Join(columnList, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updateClause, ", "))

	// Use batch insert for better performance
	// Convert all rows to UTF-8 first, then add to batch
	batch := &pgx.Batch{}
	for _, row := range rows {
		// Convert all string values from Windows-1251 to UTF-8
		convertedRow := convertRowValues(row)
		batch.Queue(query, convertedRow...)
	}

	// Execute batch
	br := tx.SendBatch(ctx, batch)

	// Check for errors in batch execution
	for i := 0; i < len(rows); i++ {
		_, err := br.Exec()
		if err != nil {
			if closeErr := br.Close(); closeErr != nil {
				return fmt.Errorf("failed to insert data into %s (row %d): %v; close batch: %w", tableName, i+1, err, closeErr)
			}
			return fmt.Errorf("failed to insert data into %s (row %d): %w", tableName, i+1, err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close batch for %s: %w", tableName, err)
	}

	slog.Info("Successfully loaded rows",
		"table", tableName,
		"rows", len(rows),
		"event", "db_load_complete",
	)
	return nil
}

// LoadTxTable loads data into a tx_* table using schema-based mapping.
func (p *Pool) LoadTxTable(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
	if data == nil {
		return nil
	}
	schema, ok := models.TxSchemas[tableName]
	if !ok {
		return fmt.Errorf("unknown tx table: %s", tableName)
	}

	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("data for %s is not a slice", tableName)
	}
	if rv.Len() == 0 {
		return nil
	}

	columns := make([]string, 0, len(schema))
	for _, spec := range schema {
		columns = append(columns, spec.Name)
	}

	rows := make([][]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		row, err := buildTxRow(schema, rv.Index(i).Interface())
		if err != nil {
			return fmt.Errorf("failed to build row for %s: %w", tableName, err)
		}
		rows[i] = row
	}

	return p.LoadData(ctx, tx, tableName, columns, rows)
}

func buildTxRow(schema []models.TxColumnSpec, value interface{}) ([]interface{}, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("row value is not struct: %T", value)
	}

	row := make([]interface{}, 0, len(schema))
	for _, spec := range schema {
		fieldName := models.ColumnToFieldName(spec.Name)
		field := rv.FieldByName(fieldName)
		if !field.IsValid() {
			return nil, fmt.Errorf("missing field %s on %s", fieldName, rv.Type().Name())
		}

		var val interface{}
		switch spec.Kind {
		case models.TxColumnString:
			val = field.String()
		case models.TxColumnInt64:
			val = field.Int()
		case models.TxColumnFloat64:
			val = field.Float()
		case models.TxColumnDate, models.TxColumnTime:
			if field.Type() != reflect.TypeOf(time.Time{}) {
				return nil, fmt.Errorf("field %s is not time.Time", fieldName)
			}
			parsed, ok := field.Interface().(time.Time)
			if !ok {
				return nil, fmt.Errorf("field %s is not time.Time", fieldName)
			}
			val = parsed
		case models.TxColumnSource:
			val = field.String()
		default:
			return nil, fmt.Errorf("unsupported column kind for %s", spec.Name)
		}

		if spec.AllowZero {
			row = append(row, safeValueAllowZero(val))
		} else {
			row = append(row, safeValue(val))
		}
	}

	return row, nil
}
