package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/models"
)

// Loader handles loading data into the database
type Loader struct {
	db *db.Pool
}

// NewLoader creates a new loader instance
func NewLoader(database *db.Pool) *Loader {
	return &Loader{
		db: database,
	}
}

// isRetryableError checks if an error is a retryable database error (deadlock, serialization failure)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL error codes for retryable errors:
		// 40001 - serialization_failure
		// 40P01 - deadlock_detected
		return pgErr.Code == "40001" || pgErr.Code == "40P01"
	}

	// Also check for error messages that might indicate deadlock
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "deadlock") || strings.Contains(errMsg, "serialization")
}

// LoadFileData loads all transaction data from a file into the database
// Implements retry logic with exponential backoff for deadlock errors
func (l *Loader) LoadFileData(ctx context.Context, transactions map[string]interface{}) error {
	// Log all transaction types found
	slog.DebugContext(ctx, "Transaction types found", "event", "transaction_type_scan")
	for tableName, data := range transactions {
		count := sliceLen(data)
		if count > 0 {
			slog.InfoContext(ctx, "Transaction type found",
				"table", tableName,
				"count", count,
				"event", "transaction_type_count",
			)
		} else if count == -1 {
			slog.WarnContext(ctx, "Unknown transaction type",
				"table", tableName,
				"type", fmt.Sprintf("%T", data),
				"event", "unknown_transaction_type",
			)
		}
	}

	// Retry logic with exponential backoff for deadlock errors
	const maxRetries = 5
	const initialBackoff = 100 * time.Millisecond
	const maxBackoff = 5 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Start a transaction for the entire file
		tx, err := l.db.BeginTx(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Load each transaction type
		loadErr := func() error {
			defer func() { _ = tx.Rollback(ctx) }() // Always rollback on error

			for tableName, data := range transactions {
				if err := l.loadTransactionType(ctx, tx, tableName, data); err != nil {
					return fmt.Errorf("failed to load %s: %w", tableName, err)
				}
			}

			// Commit the transaction
			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			return nil
		}()

		// If no error, we're done
		if loadErr == nil {
			return nil
		}

		lastErr = loadErr

		// Check if error is retryable
		if !isRetryableError(loadErr) {
			// Non-retryable error, return immediately
			return loadErr
		}

		// If this was the last attempt, return the error
		if attempt == maxRetries-1 {
			return fmt.Errorf("failed after %d retries: %w", maxRetries, loadErr)
		}

		// Calculate exponential backoff: initialBackoff * 2^attempt, capped at maxBackoff
		backoff := time.Duration(float64(initialBackoff) * math.Pow(2, float64(attempt)))
		if backoff > maxBackoff {
			backoff = maxBackoff
		}

		slog.WarnContext(ctx, "Retryable error detected, retrying",
			"attempt", attempt+1,
			"max_retries", maxRetries,
			"error", loadErr.Error(),
			"backoff", backoff.String(),
			"event", "retry_deadlock",
		)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled during retry: %w", ctx.Err())
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// loadTransactionType loads a specific transaction type
func (l *Loader) loadTransactionType(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
	if _, ok := models.TxSchemas[tableName]; !ok {
		return fmt.Errorf("unknown transaction type: %s", tableName)
	}
	return l.db.LoadTxTable(ctx, tx, tableName, data)
}

// loadTransactionRegistrations loads transaction registrations.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadTransactionRegistrations(ctx context.Context, tx pgx.Tx, transactions []models.TransactionRegistration) error {
	if len(transactions) == 0 {
		return nil
	}

	// Use the database pool's method for loading with transaction
	return l.db.LoadTransactionRegistrations(ctx, tx, transactions)
}

// loadSpecialPrices loads special prices.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadSpecialPrices(ctx context.Context, tx pgx.Tx, prices []models.SpecialPrice) error {
	if len(prices) == 0 {
		return nil
	}

	return l.db.LoadSpecialPrices(ctx, tx, prices)
}

// loadBonusTransactions loads bonus transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadBonusTransactions(ctx context.Context, tx pgx.Tx, transactions []models.BonusTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadBonusTransactions(ctx, tx, transactions)
}

// loadDiscountTransactions loads discount transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadDiscountTransactions(ctx context.Context, tx pgx.Tx, transactions []models.DiscountTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadDiscountTransactions(ctx, tx, transactions)
}

// loadBillRegistrations loads bill registrations.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadBillRegistrations(ctx context.Context, tx pgx.Tx, bills []models.BillRegistration) error {
	if len(bills) == 0 {
		return nil
	}

	return l.db.LoadBillRegistrations(ctx, tx, bills)
}

// loadEmployeeEdits loads employee edits.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadEmployeeEdits(ctx context.Context, tx pgx.Tx, edits []models.EmployeeEdit) error {
	if len(edits) == 0 {
		return nil
	}

	return l.db.LoadEmployeeEdits(ctx, tx, edits)
}

// loadEmployeeAccounting loads employee accounting.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadEmployeeAccounting(ctx context.Context, tx pgx.Tx, accounting []models.EmployeeAccounting) error {
	if len(accounting) == 0 {
		return nil
	}

	return l.db.LoadEmployeeAccounting(ctx, tx, accounting)
}

// loadVatKKTTransactions loads VAT KKT transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadVatKKTTransactions(ctx context.Context, tx pgx.Tx, transactions []models.VatKKTTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadVatKKTTransactions(ctx, tx, transactions)
}

// loadAdditionalTransactions loads additional transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadAdditionalTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AdditionalTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadAdditionalTransactions(ctx, tx, transactions)
}

// loadAstuExchangeTransactions loads ASTU exchange transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadAstuExchangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AstuExchangeTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadAstuExchangeTransactions(ctx, tx, transactions)
}

// loadCounterChangeTransactions loads counter change transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadCounterChangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.CounterChangeTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadCounterChangeTransactions(ctx, tx, transactions)
}

// loadKKTShiftReports loads KKT shift reports.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadKKTShiftReports(ctx context.Context, tx pgx.Tx, reports []models.KKTShiftReport) error {
	if len(reports) == 0 {
		return nil
	}

	return l.db.LoadKKTShiftReports(ctx, tx, reports)
}

// loadFrontolMarkUnitTransactions loads Frontol mark unit transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadFrontolMarkUnitTransactions(ctx context.Context, tx pgx.Tx, transactions []models.FrontolMarkUnitTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadFrontolMarkUnitTransactions(ctx, tx, transactions)
}

// loadBonusPayments loads bonus payments.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadBonusPayments(ctx context.Context, tx pgx.Tx, payments []models.BonusPayment) error {
	if len(payments) == 0 {
		return nil
	}

	return l.db.LoadBonusPayments(ctx, tx, payments)
}

// GetTransactionCount returns the total number of transactions
func (l *Loader) GetTransactionCount(transactions map[string]interface{}) int {
	total := 0
	for _, data := range transactions {
		if count := sliceLen(data); count > 0 {
			total += count
		}
	}
	return total
}

// GetTransactionDetails returns detailed statistics by transaction type
func (l *Loader) GetTransactionDetails(transactions map[string]interface{}) []map[string]interface{} {
	details := make([]map[string]interface{}, 0)
	for tableName, data := range transactions {
		count := sliceLen(data)
		if count > 0 {
			details = append(details, map[string]interface{}{
				"table_name": tableName,
				"count":      count,
			})
		}
	}
	return details
}

// PrintStatistics prints loading statistics using structured logging
func (l *Loader) PrintStatistics(ctx context.Context, transactions map[string]interface{}, startTime time.Time) {
	duration := time.Since(startTime)
	totalCount := l.GetTransactionCount(transactions)

	slog.InfoContext(ctx, "Loading statistics",
		"event", "loading_statistics",
		"total_processing_time", duration.String(),
		"total_transactions_loaded", totalCount,
	)

	for tableName, data := range transactions {
		count := sliceLen(data)
		if count > 0 {
			slog.DebugContext(ctx, "Table statistics",
				"table", tableName,
				"count", count,
			)
		}
	}
}

// loadFiscalPayments loads fiscal payments.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.FiscalPayment) error {
	if len(payments) == 0 {
		return nil
	}

	return l.db.LoadFiscalPayments(ctx, tx, payments)
}

// loadDocumentOperations loads document operations.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadDocumentOperations(ctx context.Context, tx pgx.Tx, operations []models.DocumentOperation) error {
	if len(operations) == 0 {
		return nil
	}

	return l.db.LoadDocumentOperations(ctx, tx, operations)
}

// loadDocumentDiscounts loads document discounts.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadDocumentDiscounts(ctx context.Context, tx pgx.Tx, discounts []models.DocumentDiscount) error {
	if len(discounts) == 0 {
		return nil
	}

	return l.db.LoadDocumentDiscounts(ctx, tx, discounts)
}

// loadCardStatusChanges loads card status changes.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadCardStatusChanges(ctx context.Context, tx pgx.Tx, changes []models.CardStatusChange) error {
	if len(changes) == 0 {
		return nil
	}

	return l.db.LoadCardStatusChanges(ctx, tx, changes)
}

// loadModifierTransactions loads modifier transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadModifierTransactions(ctx context.Context, tx pgx.Tx, transactions []models.ModifierTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadModifierTransactions(ctx, tx, transactions)
}

// loadPrepaymentTransactions loads prepayment transactions.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadPrepaymentTransactions(ctx context.Context, tx pgx.Tx, transactions []models.PrepaymentTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return l.db.LoadPrepaymentTransactions(ctx, tx, transactions)
}

// loadNonFiscalPayments loads non-fiscal payments.
//
//nolint:unused // Legacy load path retained for future wiring.
func (l *Loader) loadNonFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.NonFiscalPayment) error {
	if len(payments) == 0 {
		return nil
	}

	return l.db.LoadNonFiscalPayments(ctx, tx, payments)
}

func sliceLen(data interface{}) int {
	if data == nil {
		return 0
	}
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice {
		return -1
	}
	return rv.Len()
}

func schemaColumns(schema []models.TxColumnSpec) string {
	columns := make([]string, 0, len(schema))
	for _, spec := range schema {
		columns = append(columns, spec.Name)
	}
	return strings.Join(columns, ", ")
}

func scanTxRowInto(dst interface{}, schema []models.TxColumnSpec, values []interface{}) error {
	if len(values) < len(schema) {
		return fmt.Errorf("not enough values for tx row: got %d want %d", len(values), len(schema))
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("destination must be pointer to struct")
	}
	structVal := rv.Elem()

	for i, spec := range schema {
		fieldName := models.ColumnToFieldName(spec.Name)
		field := structVal.FieldByName(fieldName)
		if !field.IsValid() {
			return fmt.Errorf("missing field %s on %s", fieldName, structVal.Type().Name())
		}
		if !field.CanSet() {
			return fmt.Errorf("cannot set field %s on %s", fieldName, structVal.Type().Name())
		}

		val := values[i]
		if val == nil {
			continue
		}

		switch spec.Kind {
		case models.TxColumnString, models.TxColumnSource:
			switch v := val.(type) {
			case string:
				field.SetString(v)
			case []byte:
				field.SetString(string(v))
			default:
				field.SetString(fmt.Sprint(v))
			}
		case models.TxColumnInt64:
			switch v := val.(type) {
			case int64:
				field.SetInt(v)
			case int32:
				field.SetInt(int64(v))
			case int:
				field.SetInt(int64(v))
			case float64:
				field.SetInt(int64(v))
			default:
				return fmt.Errorf("unexpected int64 type for %s: %T", spec.Name, val)
			}
		case models.TxColumnFloat64:
			switch v := val.(type) {
			case float64:
				field.SetFloat(v)
			case float32:
				field.SetFloat(float64(v))
			case int64:
				field.SetFloat(float64(v))
			case int:
				field.SetFloat(float64(v))
			default:
				return fmt.Errorf("unexpected float64 type for %s: %T", spec.Name, val)
			}
		case models.TxColumnDate, models.TxColumnTime:
			t, ok := val.(time.Time)
			if !ok {
				return fmt.Errorf("unexpected time type for %s: %T", spec.Name, val)
			}
			field.Set(reflect.ValueOf(t))
		default:
			return fmt.Errorf("unsupported column kind for %s", spec.Name)
		}
	}

	return nil
}

func buildExportRow(schema []models.TxColumnSpec, values []interface{}) (ExportRow, error) {
	if len(values) < len(schema) {
		return ExportRow{}, fmt.Errorf("not enough values for export row: got %d want %d", len(values), len(schema))
	}

	row := ExportRow{}
	fields := make([]string, 0, len(schema)-1)

	for i, spec := range schema {
		val := values[i]
		if spec.Name == "source_folder" {
			if val != nil {
				row.SourceFolder = fmt.Sprint(val)
			}
			continue
		}

		if spec.Name == "transaction_id_unique" {
			if parsed, ok := toInt64(val); ok {
				row.TransactionIDUnique = parsed
			}
		}
		if spec.Name == "transaction_date" {
			if parsed, ok := toTime(val); ok {
				row.TransactionDate = parsed
			}
		}
		if spec.Name == "transaction_time" {
			if parsed, ok := toTime(val); ok {
				row.TransactionTime = parsed
			}
		}
		if spec.Name == "transaction_type" {
			if parsed, ok := toInt64(val); ok {
				row.TransactionType = int(parsed)
			}
		}

		fields = append(fields, formatTxValue(spec.Kind, val))
	}

	row.RawLine = strings.Join(fields, ";")
	return row, nil
}

func formatTxValue(kind models.TxColumnKind, val interface{}) string {
	if val == nil {
		return ""
	}
	switch kind {
	case models.TxColumnString, models.TxColumnSource:
		switch v := val.(type) {
		case string:
			return v
		case []byte:
			return string(v)
		default:
			return fmt.Sprint(v)
		}
	case models.TxColumnInt64:
		if parsed, ok := toInt64(val); ok {
			return strconv.FormatInt(parsed, 10)
		}
		return ""
	case models.TxColumnFloat64:
		if parsed, ok := toFloat64(val); ok {
			return strconv.FormatFloat(parsed, 'f', -1, 64)
		}
		return ""
	case models.TxColumnDate:
		if parsed, ok := toTime(val); ok && !parsed.IsZero() {
			return parsed.Format("02.01.2006")
		}
		return ""
	case models.TxColumnTime:
		if parsed, ok := toTime(val); ok && !parsed.IsZero() {
			return parsed.Format("15:04:05")
		}
		return ""
	default:
		return ""
	}
}

func toInt64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

func toTime(val interface{}) (time.Time, bool) {
	if val == nil {
		return time.Time{}, false
	}
	t, ok := val.(time.Time)
	return t, ok
}

// GetTransactionRegistrationsByKassaAndDate retrieves transaction registrations from database
// filtered by cash register code and transaction date
func (l *Loader) GetTransactionRegistrationsByKassaAndDate(ctx context.Context, cashRegisterCode int64, date string) ([]models.TxItemRegistration1_11, error) {
	schema := models.TxSchemas["tx_item_registration_1_11"]
	columns := schemaColumns(schema)
	query := fmt.Sprintf(`
		SELECT %s
		FROM tx_item_registration_1_11
		WHERE cash_register_code = $1 AND transaction_date = $2
		ORDER BY transaction_time, document_number
	`, columns)

	rows, err := l.db.Query(ctx, query, cashRegisterCode, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query tx_item_registration_1_11: %w", err)
	}
	defer rows.Close()

	var transactions []models.TxItemRegistration1_11
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row values: %w", err)
		}
		var t models.TxItemRegistration1_11
		if err := scanTxRowInto(&t, schema, values); err != nil {
			return nil, fmt.Errorf("failed to scan tx_item_registration_1_11: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return transactions, nil
}

// GetAvailableCashRegisters retrieves list of unique cash register codes from database
func (l *Loader) GetAvailableCashRegisters(ctx context.Context) ([]int64, error) {
	query := `
		SELECT DISTINCT cash_register_code 
		FROM tx_item_registration_1_11 
		ORDER BY cash_register_code
	`

	rows, err := l.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cash registers: %w", err)
	}
	defer rows.Close()

	var codes []int64
	for rows.Next() {
		var code int64
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("failed to scan cash register code: %w", err)
		}
		codes = append(codes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return codes, nil
}

// GetTransactionRegistrationsBySourceFolderAndDate retrieves transaction registrations from database
// filtered by source folder (kassa code from configuration) and transaction date
// sourceFolder should be in format "KASSA_CODE" or "KASSA_CODE/FOLDER_NAME"
// If only KASSA_CODE is provided, matches all folders for that kassa
func (l *Loader) GetTransactionRegistrationsBySourceFolderAndDate(ctx context.Context, sourceFolder string, date string) ([]models.TxItemRegistration1_11, error) {
	// Build WHERE clause based on source_folder format
	var whereClause string
	var args []interface{}

	if strings.Contains(sourceFolder, "/") {
		// Exact match: "P13/P13" or "N22/N22_Inter"
		whereClause = "source_folder = $1 AND transaction_date = $2"
		args = []interface{}{sourceFolder, date}
	} else {
		// Prefix match: "P13" matches "P13/P13", "P13/OtherFolder", etc.
		whereClause = "source_folder LIKE $1 AND transaction_date = $2"
		args = []interface{}{sourceFolder + "/%", date}
	}

	schema := models.TxSchemas["tx_item_registration_1_11"]
	columns := schemaColumns(schema)
	query := fmt.Sprintf(`
		SELECT %s
		FROM tx_item_registration_1_11
		WHERE %s
		ORDER BY transaction_time, document_number
	`, columns, whereClause)

	rows, err := l.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction registrations: %w", err)
	}
	defer rows.Close()

	var transactions []models.TxItemRegistration1_11
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row values: %w", err)
		}
		var t models.TxItemRegistration1_11
		if err := scanTxRowInto(&t, schema, values); err != nil {
			return nil, fmt.Errorf("failed to scan tx_item_registration_1_11: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return transactions, nil
}

// GetAvailableSourceFolders retrieves list of unique source folders from database
func (l *Loader) GetAvailableSourceFolders(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT source_folder 
		FROM tx_item_registration_1_11 
		ORDER BY source_folder
	`

	rows, err := l.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query source folders: %w", err)
	}
	defer rows.Close()

	var folders []string
	for rows.Next() {
		var folder string
		if err := rows.Scan(&folder); err != nil {
			return nil, fmt.Errorf("failed to scan source folder: %w", err)
		}
		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return folders, nil
}

// ExportRow represents a single transaction row with serialized fields.
type ExportRow struct {
	TransactionIDUnique int64
	SourceFolder        string
	TransactionDate     time.Time
	TransactionTime     time.Time
	TransactionType     int
	RawLine             string
}

// GetAllTransactionsBySourceFolderAndDate retrieves ALL transaction types from ALL tables
// ordered by transaction_time, transaction_id_unique
func (l *Loader) GetAllTransactionsBySourceFolderAndDate(ctx context.Context, sourceFolder string, date string) ([]ExportRow, error) {
	slog.InfoContext(ctx, "GetAllTransactionsBySourceFolderAndDate called",
		"source_folder", sourceFolder,
		"date", date,
		"event", "get_all_transactions_start",
	)

	// Build WHERE condition and args based on source_folder format
	var whereCondition string
	var args []interface{}

	if strings.Contains(sourceFolder, "/") {
		// Exact match: "P13/P13" or "N22/N22_Inter"
		whereCondition = "source_folder = $1 AND transaction_date = $2"
		args = []interface{}{sourceFolder, date}
		slog.DebugContext(ctx, "Using exact match for source_folder",
			"where_condition", whereCondition,
			"args", args,
		)
	} else {
		// Prefix match: "P13" matches "P13/P13", "P13/OtherFolder", etc.
		whereCondition = "source_folder LIKE $1 AND transaction_date = $2"
		args = []interface{}{sourceFolder + "/%", date}
		slog.DebugContext(ctx, "Using prefix match for source_folder",
			"where_condition", whereCondition,
			"args", args,
		)
	}

	tables := make([]string, 0, len(models.TxSchemas))
	for name := range models.TxSchemas {
		tables = append(tables, name)
	}
	sort.Strings(tables)

	slog.DebugContext(ctx, "Building UNION ALL query",
		"table_count", len(tables),
		"tables", tables,
	)

	var transactions []ExportRow
	transactionTypesCount := make(map[int]int)

	for _, table := range tables {
		schema := models.TxSchemas[table]
		query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", schemaColumns(schema), table, whereCondition)
		rows, err := l.db.Query(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to query %s: %w", table, err)
		}

		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to read values from %s: %w", table, err)
			}
			row, err := buildExportRow(schema, values)
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to build export row for %s: %w", table, err)
			}
			transactions = append(transactions, row)
			transactionTypesCount[row.TransactionType]++
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows for %s: %w", table, err)
		}
	}

	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].TransactionTime.Equal(transactions[j].TransactionTime) {
			return transactions[i].TransactionIDUnique < transactions[j].TransactionIDUnique
		}
		return transactions[i].TransactionTime.Before(transactions[j].TransactionTime)
	})

	slog.DebugContext(ctx, "Transaction type distribution",
		"type_counts", transactionTypesCount,
		"event", "transaction_type_distribution",
	)

	slog.InfoContext(ctx, "GetAllTransactionsBySourceFolderAndDate completed",
		"source_folder", sourceFolder,
		"date", date,
		"transactions_found", len(transactions),
		"event", "get_all_transactions_complete",
	)

	return transactions, nil
}
