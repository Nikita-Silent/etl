package db

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

// LoadTransactionRegistrations loads transaction registrations
func (p *Pool) LoadTransactionRegistrations(ctx context.Context, tx pgx.Tx, transactions []models.TransactionRegistration) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "item_code", "group_code",
		"amount_total", "quantity", "amount_cash_register", "operation_type", "shift_number",
		"item_price", "item_sum", "print_group_code", "item_line_number", "article_sku", "registration_barcode",
		"position_amount", "kkt_section", "reserved_field22", "document_type_code",
		"comment_code", "reserved_field25", "document_info", "enterprise_id", "employee_code",
		"divided_pack_qty", "gift_card_number", "pack_quantity", "nomenclature_type",
		"marking_code", "excise_stamp", "personal_mod_group", "lottery_time", "lottery_id",
		"reserved_field38", "alc_code", "reserved_field40", "prescription_data1",
		"prescription_data2", "coupons_per_item", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.TransactionIDUnique, t.SourceFolder, t.TransactionDate, t.TransactionTime, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.ItemCode, t.GroupCode,
			t.AmountTotal, t.Quantity, t.AmountCashRegister, t.OperationType, t.ShiftNumber,
			t.ItemPrice, t.ItemSum, t.PrintGroupCode, t.ItemLineNumber, t.ArticleSKU, t.RegistrationBarcode,
			t.PositionAmount, t.KKTSection, t.ReservedField22, t.DocumentTypeCode,
			t.CommentCode, t.ReservedField25, t.DocumentInfo, t.EnterpriseID, t.EmployeeCode,
			t.DividedPackQty, t.GiftCardNumber, t.PackQuantity, t.NomenclatureType,
			t.MarkingCode, t.ExciseStamp, t.PersonalModGroup, t.LotteryTime, t.LotteryID,
			t.ReservedField38, t.ALCCode, t.ReservedField40, t.PrescriptionData1,
			t.PrescriptionData2, t.CouponsPerItem, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "transaction_registrations", columns, rows)
}

// LoadSpecialPrices loads special prices
func (p *Pool) LoadSpecialPrices(ctx context.Context, tx pgx.Tx, prices []models.SpecialPrice) error {
	if len(prices) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"price_list_code", "group_code", "price_type", "special_price", "product_card_price",
		"operation_type", "promotion_code", "event_code", "print_group_code",
		"document_type_code", "document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(prices))
	for i, p := range prices {
		rows[i] = []interface{}{
			p.ID, p.SourceFolder, p.Date, p.Time, p.TransactionType,
			p.CashRegisterCode, p.DocumentNumber, p.CashierCode, p.ShiftNumber,
			p.PriceListCode, p.GroupCode, p.PriceType, p.SpecialPrice, p.ProductCardPrice,
			p.OperationType, p.PromotionCode, p.EventCode, p.PrintGroupCode,
			p.DocumentTypeCode, p.DocumentInfo, p.EnterpriseID, p.RawData,
		}
	}

	return p.LoadData(ctx, tx, "special_prices", columns, rows)
}

// LoadBonusTransactions loads bonus transactions
func (p *Pool) LoadBonusTransactions(ctx context.Context, tx pgx.Tx, transactions []models.BonusTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"bonus_amount", "accrued_bonus_amount", "bonus_type", "operation_type",
		"promotion_code", "event_code", "print_group_code", "document_type_code",
		"document_info", "enterprise_id", "ps_protocol_number", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.ID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.ShiftNumber,
			t.BonusAmount, t.AccruedBonusAmount, t.BonusType, t.OperationType,
			t.PromotionCode, t.EventCode, t.PrintGroupCode, t.DocumentTypeCode,
			t.DocumentInfo, t.EnterpriseID, t.PSProtocolNumber, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "bonus_transactions", columns, rows)
}

// LoadDiscountTransactions loads discount transactions
func (p *Pool) LoadDiscountTransactions(ctx context.Context, tx pgx.Tx, transactions []models.DiscountTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"discount_amount", "discount_type", "discount_value", "discount_percent", "operation_type",
		"promotion_code", "event_code", "print_group_code", "document_type_code",
		"document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.ID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.ShiftNumber,
			t.DiscountAmount, t.DiscountType, t.DiscountValue, t.DiscountPercent, t.OperationType,
			t.PromotionCode, t.EventCode, t.PrintGroupCode, t.DocumentTypeCode,
			t.DocumentInfo, t.EnterpriseID, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "discount_transactions", columns, rows)
}

// LoadBillRegistrations loads bill registrations
func (p *Pool) LoadBillRegistrations(ctx context.Context, tx pgx.Tx, bills []models.BillRegistration) error {
	if len(bills) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "bill_code",
		"group_code", "bill_denomination", "bill_quantity", "bill_amount",
		"operation_type", "shift_number", "bill_number", "bill_total_amount",
		"bill_type", "print_group_code", "customer_code", "reserved_field20",
		"reserved_field21", "reserved_field22", "document_type_code",
		"reserved_field24", "reserved_field25", "raw_data",
	}

	rows := make([][]interface{}, len(bills))
	for i, b := range bills {
		// Convert Date (time.Time) to string format YYYY-MM-DD for transaction_date
		transactionDate := b.Date.Format("2006-01-02")
		// Convert Time (time.Time) to string format HH:MM:SS for transaction_time
		transactionTime := b.Time.Format("15:04:05")

		rows[i] = []interface{}{
			b.ID, b.SourceFolder, transactionDate, transactionTime, b.TransactionType,
			b.CashRegisterCode, b.DocumentNumber, b.CashierCode, b.BillCode,
			b.GroupCode, b.BillDenomination, b.BillQuantity, b.BillAmount,
			b.OperationType, b.ShiftNumber, b.BillNumber, b.BillTotalAmount,
			b.BillType, b.PrintGroupCode, b.CustomerCode, b.ReservedField20,
			b.ReservedField21, b.ReservedField22, b.DocumentTypeCode,
			b.ReservedField24, b.ReservedField25, safeValue(b.RawData),
		}
	}

	return p.LoadData(ctx, tx, "bill_registrations", columns, rows)
}

// LoadEmployeeEdits loads employee edits
func (p *Pool) LoadEmployeeEdits(ctx context.Context, tx pgx.Tx, edits []models.EmployeeEdit) error {
	if len(edits) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"employee_code", "operation_type", "document_type_code", "document_info",
		"enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(edits))
	for i, e := range edits {
		rows[i] = []interface{}{
			e.ID, e.SourceFolder, e.Date, e.Time, e.TransactionType,
			e.CashRegisterCode, e.DocumentNumber, e.CashierCode, e.ShiftNumber,
			e.EmployeeCode, e.OperationType, e.DocumentTypeCode, e.DocumentInfo,
			e.EnterpriseID, e.RawData,
		}
	}

	return p.LoadData(ctx, tx, "employee_edits", columns, rows)
}

// LoadEmployeeAccounting loads employee accounting
func (p *Pool) LoadEmployeeAccounting(ctx context.Context, tx pgx.Tx, accounting []models.EmployeeAccounting) error {
	if len(accounting) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"employee_code", "operation_type", "print_group_code", "document_type_code",
		"document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(accounting))
	for i, a := range accounting {
		rows[i] = []interface{}{
			a.ID, a.SourceFolder, a.Date, a.Time, a.TransactionType,
			a.CashRegisterCode, a.DocumentNumber, a.CashierCode, a.ShiftNumber,
			a.EmployeeCode, a.OperationType, a.PrintGroupCode, a.DocumentTypeCode,
			a.DocumentInfo, a.EnterpriseID, a.RawData,
		}
	}

	return p.LoadData(ctx, tx, "employee_accounting", columns, rows)
}

// LoadVatKKTTransactions loads VAT KKT transactions
func (p *Pool) LoadVatKKTTransactions(ctx context.Context, tx pgx.Tx, transactions []models.VatKKTTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "operation_type",
		"shift_number", "print_group_code", "document_type_code", "document_info", "enterprise_id",
		"vat_0_amount", "vat_10_amount", "vat_20_amount", "amount_without_vat",
		"vat_10_110_amount", "vat_20_120_amount", "reserved_fields", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.ID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.OperationType,
			t.ShiftNumber, t.PrintGroupCode, t.DocumentTypeCode, t.DocumentInfo, t.EnterpriseID,
			t.Vat0Amount, t.Vat10Amount, t.Vat20Amount, t.AmountWithoutVat,
			t.Vat10_110Amount, t.Vat20_120Amount, t.ReservedFields, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "vat_kkt_transactions", columns, rows)
}

// LoadAdditionalTransactions loads additional transactions
func (p *Pool) LoadAdditionalTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AdditionalTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"additional_type", "additional_amount", "additional_info", "operation_type",
		"print_group_code", "document_type_code", "document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.ID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.ShiftNumber,
			t.AdditionalType, t.AdditionalAmount, t.AdditionalInfo, t.OperationType,
			t.PrintGroupCode, t.DocumentTypeCode, t.DocumentInfo, t.EnterpriseID, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "additional_transactions", columns, rows)
}

// LoadAstuExchangeTransactions loads ASTU exchange transactions
func (p *Pool) LoadAstuExchangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.AstuExchangeTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"exchange_type", "exchange_amount", "exchange_rate", "operation_type",
		"print_group_code", "document_type_code", "document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.ID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.ShiftNumber,
			t.ExchangeType, t.ExchangeAmount, t.ExchangeRate, t.OperationType,
			t.PrintGroupCode, t.DocumentTypeCode, t.DocumentInfo, t.EnterpriseID, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "astu_exchange_transactions", columns, rows)
}

// LoadCounterChangeTransactions loads counter change transactions
func (p *Pool) LoadCounterChangeTransactions(ctx context.Context, tx pgx.Tx, transactions []models.CounterChangeTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"card_number_client_code", "card_type_code", "binding_type", "value_after_changes",
		"change_amount", "operation_type", "promotion_code", "event_code",
		"counter_type_code", "counter_code", "document_type_code", "document_info",
		"enterprise_id", "counter_movement_start_date", "card_validity_start_date",
		"card_validity_end_date", "counter_movement_end_date", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.ID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.ShiftNumber,
			t.CardNumberOrClientCode, t.CardTypeCode, t.BindingType, t.ValueAfterChanges,
			t.ChangeAmount, t.OperationType, t.PromotionCode, t.EventCode,
			t.CounterTypeCode, t.CounterCode, t.DocumentTypeCode, t.DocumentInfo,
			t.EnterpriseID, t.MovementStartDate, t.CardStartDate,
			t.CardEndDate, t.MovementEndDate, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "counter_change_transactions", columns, rows)
}

// LoadKKTShiftReports loads KKT shift reports
func (p *Pool) LoadKKTShiftReports(ctx context.Context, tx pgx.Tx, reports []models.KKTShiftReport) error {
	if len(reports) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"report_type", "report_data", "report_amount", "operation_type",
		"print_group_code", "document_type_code", "document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(reports))
	for i, r := range reports {
		rows[i] = []interface{}{
			r.ID, r.SourceFolder, r.Date, r.Time, r.TransactionType,
			r.CashRegisterCode, r.DocumentNumber, r.CashierCode, r.ShiftNumber,
			r.ReportType, r.ReportData, r.ReportAmount, r.OperationType,
			r.PrintGroupCode, r.DocumentTypeCode, r.DocumentInfo, r.EnterpriseID, r.RawData,
		}
	}

	return p.LoadData(ctx, tx, "kkt_shift_reports", columns, rows)
}

// LoadFrontolMarkUnitTransactions loads Frontol mark unit transactions
func (p *Pool) LoadFrontolMarkUnitTransactions(ctx context.Context, tx pgx.Tx, transactions []models.FrontolMarkUnitTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"mark_unit_type", "mark_unit_code", "mark_unit_data", "operation_type",
		"print_group_code", "document_type_code", "document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		rows[i] = []interface{}{
			t.ID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, t.DocumentNumber, t.CashierCode, t.ShiftNumber,
			t.MarkUnitType, t.MarkUnitCode, t.MarkUnitData, t.OperationType,
			t.PrintGroupCode, t.DocumentTypeCode, t.DocumentInfo, t.EnterpriseID, t.RawData,
		}
	}

	return p.LoadData(ctx, tx, "frontol_mark_unit_transactions", columns, rows)
}

// LoadBonusPayments loads bonus payments
func (p *Pool) LoadBonusPayments(ctx context.Context, tx pgx.Tx, payments []models.BonusPayment) error {
	if len(payments) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "shift_number",
		"bonus_card_number", "payment_type", "counter_change_value", "payment_amount", "operation_type",
		"promotion_code", "event_code", "print_group_code", "document_type_code",
		"document_info", "enterprise_id", "ps_protocol_number", "raw_data",
	}

	rows := make([][]interface{}, len(payments))
	for i, p := range payments {
		rows[i] = []interface{}{
			p.ID, p.SourceFolder, p.Date, p.Time, p.TransactionType,
			p.CashRegisterCode, p.DocumentNumber, p.CashierCode, p.ShiftNumber,
			p.BonusCardNumber, p.PaymentType, p.CounterChangeValue, p.PaymentAmount, p.OperationType,
			p.PromotionCode, p.EventCode, p.PrintGroupCode, p.DocumentTypeCode,
			p.DocumentInfo, p.EnterpriseID, p.PSProtocolNumber, p.RawData,
		}
	}

	return p.LoadData(ctx, tx, "bonus_payments", columns, rows)
}

// LoadDocumentOperations loads document operations
func (p *Pool) LoadDocumentOperations(ctx context.Context, tx pgx.Tx, operations []models.DocumentOperation) error {
	if len(operations) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "customer_card_numbers", "dimension_value_codes",
		"reserved_field10", "quantity", "total_amount", "operation_type", "shift_number", "customer_code",
		"reserved_field16", "document_print_group_code", "bonus_amount", "order_id", "document_amount_without_discounts", "visitor_count",
		"correction_type", "kkt_registration_number", "document_type_code", "comment_code", "base_document_number",
		"employee_code", "employee_list_edit_document_number", "department_code", "hall_code", "service_point_code",
		"reservation_id", "user_variable_value", "external_comment", "revaluation_date_time", "contractor_code",
		"department_id", "reserved_field39", "coupons_on_document", "calculation_date_time", "document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(operations))
	for i, o := range operations {
		// Convert ID from string to int64
		transactionID, _ := strconv.ParseInt(o.ID, 10, 64)
		// Convert DocumentNumber from string to int64
		docNumber, _ := strconv.ParseInt(o.DocumentNumber, 10, 64)
		// Convert CashierCode from string to int64
		cashierCode, _ := strconv.ParseInt(o.CashierCode, 10, 64)

		rows[i] = []interface{}{
			transactionID, o.SourceFolder, o.Date, o.Time, o.TransactionType,
			o.CashRegisterCode, docNumber, cashierCode,
			safeValue(o.CustomerCardNumbers), safeValue(o.DimensionValueCodes),
			o.ReservedField10, // Field 10: – (empty, for compliance)
			o.Quantity, o.TotalAmount, o.OperationType, o.ShiftNumber, o.CustomerCode,
			o.ReservedField16, // Field 16: – (empty, for compliance)
			o.DocumentPrintGroupCode, safeValue(o.BonusAmount), safeValue(o.OrderID), o.DocumentAmountWithoutDiscounts, o.VisitorCount,
			o.CorrectionType, o.KKTRegistrationNumber, o.DocumentTypeCode, o.CommentCode, o.BaseDocumentNumber,
			o.EmployeeCode, o.EmployeeListEditDocumentNumber, safeValue(o.DepartmentCode), o.HallCode, o.ServicePointCode,
			safeValue(o.ReservationID), safeValue(o.UserVariableValue), safeValue(o.ExternalComment), safeValue(o.RevaluationDateTime), o.ContractorCode,
			safeValue(o.DepartmentID), o.ReservedField39, // Field 39: – (empty, for compliance)
			safeValue(o.CouponsOnDocument), safeValue(o.CalculationDateTime), safeValue(o.DocumentInfo), o.EnterpriseID, safeValue(o.RawData),
		}
	}

	return p.LoadData(ctx, tx, "document_operations", columns, rows)
}

// LoadDocumentDiscounts loads document discounts
func (p *Pool) LoadDocumentDiscounts(ctx context.Context, tx pgx.Tx, discounts []models.DocumentDiscount) error {
	if len(discounts) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "discount_info", "group_code",
		"discount_type", "discount_value", "discount_amount", "operation_type", "shift_number",
		"promotion_code", "event_code", "print_group_code", "document_type_code", "document_info",
		"enterprise_id", "reserved_fields", "raw_data",
	}

	rows := make([][]interface{}, len(discounts))
	for i, d := range discounts {
		// Convert ID from string to int64
		transactionID, _ := strconv.ParseInt(d.ID, 10, 64)
		// Convert DocumentNumber from string to int64
		docNumber, _ := strconv.ParseInt(d.DocumentNumber, 10, 64)
		// Convert CashierCode from string to int64
		cashierCode, _ := strconv.ParseInt(d.CashierCode, 10, 64)

		rows[i] = []interface{}{
			transactionID, d.SourceFolder, d.Date, d.Time, d.TransactionType,
			d.CashRegisterCode, docNumber, cashierCode,
			safeValue(d.DiscountInfo), safeValue(""), // GroupCode not in DocumentDiscount model
			d.DiscountType, d.DiscountValue, d.DiscountAmount, d.OperationType, d.ShiftNumber,
			d.CampaignCode, d.EventCode, d.PrintGroupCode, d.DocumentTypeCode, safeValue(d.DocumentInfo),
			d.EnterpriseID, nil, // reserved_fields as NULL for now
			safeValue(d.RawData),
		}
	}

	return p.LoadData(ctx, tx, "document_discounts", columns, rows)
}

// LoadFiscalPayments loads fiscal payments
func (p *Pool) LoadFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.FiscalPayment) error {
	if len(payments) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "card_number", "payment_type_code",
		"payment_type_operation", "client_amount_payment_currency", "client_amount_base_currency", "operation_type", "shift_number",
		"promotion_code", "event_code", "current_print_group_code", "currency_code", "cash_withdrawal_amount",
		"counter_type_code", "counter_code", "document_type_code", "document_info", "enterprise_id",
		"ps_protocol_number", "raw_data",
	}

	rows := make([][]interface{}, len(payments))
	for i, payment := range payments {
		// Convert ID from string to int64
		transactionID, _ := strconv.ParseInt(payment.ID, 10, 64)
		// Convert DocumentNumber from string to int64
		docNumber, _ := strconv.ParseInt(payment.DocumentNumber, 10, 64)
		// Convert CashierCode from string to int64
		cashierCode, _ := strconv.ParseInt(payment.CashierCode, 10, 64)

		rows[i] = []interface{}{
			transactionID, payment.SourceFolder, payment.Date, payment.Time, payment.TransactionType,
			payment.CashRegisterCode, docNumber, cashierCode,
			safeValue(payment.CardNumber), safeValue(payment.PaymentTypeCode),
			payment.PaymentTypeOperation, payment.CustomerAmountInPaymentCurrency, payment.CustomerAmountInBaseCurrency, payment.OperationType, payment.ShiftNumber,
			payment.PromotionCode, payment.EventCode, payment.CurrentPrintGroupCode, payment.CurrencyCode, payment.CashOutAmount,
			payment.CounterTypeCode, payment.CounterCode, payment.DocumentTypeCode, safeValue(payment.DocumentInfo), payment.EnterpriseID,
			payment.PSProtocolNumber, safeValue(payment.RawData),
		}
	}

	return p.LoadData(ctx, tx, "fiscal_payments", columns, rows)
}

// LoadCardStatusChanges loads card status changes
func (p *Pool) LoadCardStatusChanges(ctx context.Context, tx pgx.Tx, changes []models.CardStatusChange) error {
	if len(changes) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "card_number", "card_type_code",
		"card_type", "operation_type", "shift_number", "promotion_code", "event_code",
		"document_type_code", "document_info", "enterprise_id", "old_card_status", "new_card_status",
		"new_card_start_date", "new_card_end_date", "raw_data",
	}

	rows := make([][]interface{}, len(changes))
	for i, c := range changes {
		transactionID, _ := strconv.ParseInt(c.ID, 10, 64)
		docNumber, _ := strconv.ParseInt(c.DocumentNumber, 10, 64)
		cashierCode, _ := strconv.ParseInt(c.CashierCode, 10, 64)

		rows[i] = []interface{}{
			transactionID, c.SourceFolder, c.Date, c.Time, c.TransactionType,
			c.CashRegisterCode, docNumber, cashierCode,
			safeValue(c.CardNumber), safeValue(c.CardTypeCode),
			c.CardType, c.OperationType, c.ShiftNumber, c.CampaignCode, c.EventCode,
			c.DocumentTypeCode, safeValue(c.DocumentInfo), c.EnterpriseID, c.OldStatus, c.NewStatus,
			safeValue(c.NewStartDate), safeValue(c.NewEndDate), safeValue(c.RawData),
		}
	}

	return p.LoadData(ctx, tx, "card_status_changes", columns, rows)
}

// LoadModifierTransactions loads modifier transactions
func (p *Pool) LoadModifierTransactions(ctx context.Context, tx pgx.Tx, transactions []models.ModifierTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "product_identifier", "group_code",
		"product_quantity", "operation_type", "shift_number", "document_print_group_code", "document_type_code",
		"document_info", "enterprise_id", "modifier_code", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		transactionID, _ := strconv.ParseInt(t.ID, 10, 64)
		docNumber, _ := strconv.ParseInt(t.DocumentNumber, 10, 64)
		cashierCode, _ := strconv.ParseInt(t.CashierCode, 10, 64)

		rows[i] = []interface{}{
			transactionID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, docNumber, cashierCode,
			safeValue(t.ItemID), safeValue(""), // GroupCode
			t.Quantity, t.OperationType, t.ShiftNumber, t.DocumentPrintGroupCode, t.DocumentTypeCode,
			safeValue(t.DocumentInfo), t.EnterpriseID, safeValue(t.ModifierCode), safeValue(t.RawData),
		}
	}

	return p.LoadData(ctx, tx, "modifier_transactions", columns, rows)
}

// LoadPrepaymentTransactions loads prepayment transactions
func (p *Pool) LoadPrepaymentTransactions(ctx context.Context, tx pgx.Tx, transactions []models.PrepaymentTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "group_code", "prepayment_type",
		"prepayment_amount", "operation_type", "shift_number", "print_group_code", "document_type_code",
		"document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		transactionID, _ := strconv.ParseInt(t.ID, 10, 64)
		docNumber, _ := strconv.ParseInt(t.DocumentNumber, 10, 64)
		cashierCode, _ := strconv.ParseInt(t.CashierCode, 10, 64)

		rows[i] = []interface{}{
			transactionID, t.SourceFolder, t.Date, t.Time, t.TransactionType,
			t.CashRegisterCode, docNumber, cashierCode,
			safeValue(""), // GroupCode
			t.PrepaymentType, t.Amount, t.OperationType, t.ShiftNumber, t.PrintGroupCode, t.DocumentTypeCode,
			safeValue(t.DocumentInfo), t.EnterpriseID, safeValue(t.RawData),
		}
	}

	return p.LoadData(ctx, tx, "prepayment_transactions", columns, rows)
}

// LoadNonFiscalPayments loads non-fiscal payments
func (p *Pool) LoadNonFiscalPayments(ctx context.Context, tx pgx.Tx, payments []models.NonFiscalPayment) error {
	if len(payments) == 0 {
		return nil
	}

	columns := []string{
		"transaction_id_unique", "source_folder", "transaction_date", "transaction_time", "transaction_type",
		"cash_register_code", "document_number", "cashier_code", "gift_card_number", "payment_type_code",
		"payment_type_operation", "payment_amount", "operation_type", "shift_number", "promotion_code",
		"event_code", "print_group_code", "counter_type_code", "counter_code", "document_type_code",
		"document_info", "enterprise_id", "raw_data",
	}

	rows := make([][]interface{}, len(payments))
	for i, p := range payments {
		transactionID, _ := strconv.ParseInt(p.ID, 10, 64)
		docNumber, _ := strconv.ParseInt(p.DocumentNumber, 10, 64)
		cashierCode, _ := strconv.ParseInt(p.CashierCode, 10, 64)

		rows[i] = []interface{}{
			transactionID, p.SourceFolder, p.Date, p.Time, p.TransactionType,
			p.CashRegisterCode, docNumber, cashierCode,
			safeValue(p.GiftCardNumber), safeValue(p.PaymentTypeCode),
			p.PaymentTypeOperation, p.Amount, p.OperationType, p.ShiftNumber, p.CampaignCode,
			p.EventCode, p.PositionPrintGroupCode, p.CounterTypeCode, p.CounterCode, p.DocumentTypeCode,
			safeValue(p.DocumentInfo), p.EnterpriseID, safeValue(p.RawData),
		}
	}

	return p.LoadData(ctx, tx, "non_fiscal_payments", columns, rows)
}
