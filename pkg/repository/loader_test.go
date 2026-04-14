package repository

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/user/go-frontol-loader/pkg/models"
)

type fakeTx struct {
	commitCalls   int
	rollbackCalls int
	execSQL       []string
}

func (f *fakeTx) Begin(ctx context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakeTx) Commit(ctx context.Context) error {
	f.commitCalls++
	return nil
}
func (f *fakeTx) Rollback(ctx context.Context) error {
	f.rollbackCalls++
	return nil
}
func (f *fakeTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	f.execSQL = append(f.execSQL, sql)
	return pgconn.CommandTag{}, nil
}
func (f *fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (f *fakeTx) Conn() *pgx.Conn                                               { return nil }

type loaderDBMock struct {
	beginTxFunc     func(ctx context.Context) (pgx.Tx, error)
	loadTxTableFunc func(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error
}

func (m *loaderDBMock) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.beginTxFunc != nil {
		return m.beginTxFunc(ctx)
	}
	return &fakeTx{}, nil
}
func (m *loaderDBMock) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (m *loaderDBMock) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}
func (m *loaderDBMock) LoadTxTable(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
	if m.loadTxTableFunc != nil {
		return m.loadTxTableFunc(ctx, tx, tableName, data)
	}
	return nil
}

// TestIsRetryableError tests the retryable error detection.
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name: "serialization failure",
			err: &pgconn.PgError{
				Code: "40001",
			},
			expected: true,
		},
		{
			name: "deadlock detected",
			err: &pgconn.PgError{
				Code: "40P01",
			},
			expected: true,
		},
		{
			name: "other PostgreSQL error",
			err: &pgconn.PgError{
				Code: "23505",
			},
			expected: false,
		},
		{
			name:     "error with deadlock in message",
			err:      &testError{msg: "deadlock detected"},
			expected: true,
		},
		{
			name:     "error with serialization in message",
			err:      &testError{msg: "serialization failure occurred"},
			expected: true,
		},
		{
			name:     "regular error",
			err:      &testError{msg: "connection refused"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// testError is a simple error implementation for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestGetTransactionCount tests transaction counting.
func TestGetTransactionCount(t *testing.T) {
	loader := &Loader{db: nil}

	tests := []struct {
		name         string
		transactions map[string]interface{}
		expected     int
	}{
		{
			name:         "empty transactions",
			transactions: map[string]interface{}{},
			expected:     0,
		},
		{
			name: "single transaction type",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}, {TransactionIDUnique: 2}},
			},
			expected: 2,
		},
		{
			name: "multiple transaction types",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}},
				"tx_special_price_3":        []models.TxSpecialPrice3{{TransactionIDUnique: 10}, {TransactionIDUnique: 11}},
				"tx_bonus_accrual_9":        []models.TxBonusAccrual9{{TransactionIDUnique: 20}},
			},
			expected: 4,
		},
		{
			name: "unknown type (counted as slice)",
			transactions: map[string]interface{}{
				"unknown_type": []string{"test"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.GetTransactionCount(tt.transactions)
			if result != tt.expected {
				t.Errorf("GetTransactionCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestGetTransactionDetails tests transaction details extraction.
func TestGetTransactionDetails(t *testing.T) {
	loader := &Loader{db: nil}

	tests := []struct {
		name         string
		transactions map[string]interface{}
		expectedLen  int
		expectedSum  int
	}{
		{
			name:         "empty transactions",
			transactions: map[string]interface{}{},
			expectedLen:  0,
			expectedSum:  0,
		},
		{
			name: "single transaction type",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}, {TransactionIDUnique: 2}},
			},
			expectedLen: 1,
			expectedSum: 2,
		},
		{
			name: "multiple transaction types",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}},
				"tx_bonus_accrual_9":        []models.TxBonusAccrual9{{TransactionIDUnique: 10}, {TransactionIDUnique: 11}},
				"tx_document_open_42":       []models.TxDocumentOpen42{{TransactionIDUnique: 20}},
			},
			expectedLen: 3,
			expectedSum: 4,
		},
		{
			name: "empty transaction type (ignored)",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{},
			},
			expectedLen: 0,
			expectedSum: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.GetTransactionDetails(tt.transactions)
			if len(result) != tt.expectedLen {
				t.Errorf("GetTransactionDetails() length = %d, want %d", len(result), tt.expectedLen)
			}

			sum := 0
			for _, detail := range result {
				if count, ok := detail["count"].(int); ok {
					sum += count
				}
			}
			if sum != tt.expectedSum {
				t.Errorf("GetTransactionDetails() total count = %d, want %d", sum, tt.expectedSum)
			}
		})
	}
}

func TestGetTransactionDetailsSortedByTableName(t *testing.T) {
	loader := &Loader{db: nil}
	transactions := map[string]interface{}{
		"tx_special_price_3":        []models.TxSpecialPrice3{{TransactionIDUnique: 10}},
		"tx_bonus_accrual_9":        []models.TxBonusAccrual9{{TransactionIDUnique: 20}},
		"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}},
	}

	details := loader.GetTransactionDetails(transactions)
	gotOrder := []string{
		details[0]["table_name"].(string),
		details[1]["table_name"].(string),
		details[2]["table_name"].(string),
	}
	wantOrder := []string{"tx_bonus_accrual_9", "tx_item_registration_1_11", "tx_special_price_3"}
	for i := range wantOrder {
		if gotOrder[i] != wantOrder[i] {
			t.Fatalf("GetTransactionDetails() order = %v, want %v", gotOrder, wantOrder)
		}
	}
}

// TestGetTransactionCountEdgeCases tests edge cases for GetTransactionCount.
func TestGetTransactionCountEdgeCases(t *testing.T) {
	loader := &Loader{db: nil}

	tests := []struct {
		name         string
		transactions map[string]interface{}
		expected     int
	}{
		{
			name: "empty slice",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{},
			},
			expected: 0,
		},
		{
			name: "nil slice",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11(nil),
			},
			expected: 0,
		},
		{
			name: "large slice",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": make([]models.TxItemRegistration1_11, 10000),
			},
			expected: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.GetTransactionCount(tt.transactions)
			if result != tt.expected {
				t.Errorf("GetTransactionCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   int64
		wantOK bool
	}{
		{name: "int64", input: int64(10), want: 10, wantOK: true},
		{name: "int32", input: int32(5), want: 5, wantOK: true},
		{name: "int", input: int(7), want: 7, wantOK: true},
		{name: "float64", input: float64(3.9), want: 3, wantOK: true},
		{name: "float32", input: float32(2.1), want: 2, wantOK: true},
		{name: "string", input: "1", want: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toInt64(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("toInt64() ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("toInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   float64
		wantOK bool
	}{
		{name: "float64", input: float64(1.5), want: 1.5, wantOK: true},
		{name: "float32", input: float32(2.25), want: 2.25, wantOK: true},
		{name: "int64", input: int64(9), want: 9, wantOK: true},
		{name: "int", input: int(3), want: 3, wantOK: true},
		{name: "string", input: "2", want: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("toFloat64() ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("toFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		input  interface{}
		want   time.Time
		wantOK bool
	}{
		{name: "time", input: now, want: now, wantOK: true},
		{name: "nil", input: nil, want: time.Time{}, wantOK: false},
		{name: "string", input: "2024-01-01", want: time.Time{}, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toTime(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("toTime() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && !got.Equal(tt.want) {
				t.Fatalf("toTime() = %v, want %v", got, tt.want)
			}
			if !ok && !got.IsZero() {
				t.Fatalf("toTime() = %v, want zero time", got)
			}
		})
	}
}

func TestFormatTxValue(t *testing.T) {
	nonZeroDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	nonZeroTime := time.Date(2024, 3, 15, 12, 34, 56, 0, time.UTC)

	tests := []struct {
		name string
		kind models.TxColumnKind
		val  interface{}
		want string
	}{
		{name: "nil_value", kind: models.TxColumnString, val: nil, want: ""},
		{name: "string_value", kind: models.TxColumnString, val: "hello", want: "hello"},
		{name: "byte_string", kind: models.TxColumnString, val: []byte("bytes"), want: "bytes"},
		{name: "int64_value", kind: models.TxColumnInt64, val: int64(5), want: "5"},
		{name: "int_value", kind: models.TxColumnInt64, val: int(7), want: "7"},
		{name: "float_value", kind: models.TxColumnFloat64, val: float64(1.5), want: "1.5"},
		{name: "float_from_int", kind: models.TxColumnFloat64, val: int(2), want: "2"},
		{name: "nan", kind: models.TxColumnFloat64, val: math.NaN(), want: "NaN"},
		{name: "pos_inf", kind: models.TxColumnFloat64, val: math.Inf(1), want: "+Inf"},
		{name: "neg_inf", kind: models.TxColumnFloat64, val: math.Inf(-1), want: "-Inf"},
		{name: "date_value", kind: models.TxColumnDate, val: nonZeroDate, want: "15.03.2024"},
		{name: "zero_date", kind: models.TxColumnDate, val: time.Time{}, want: ""},
		{name: "time_value", kind: models.TxColumnTime, val: nonZeroTime, want: "12:34:56"},
		{name: "zero_time", kind: models.TxColumnTime, val: time.Time{}, want: ""},
		{name: "wrong_type", kind: models.TxColumnTime, val: "12:00:00", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTxValue(tt.kind, tt.val); got != tt.want {
				t.Fatalf("formatTxValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSliceLen(t *testing.T) {
	var nilSlice []int

	tests := []struct {
		name string
		val  interface{}
		want int
	}{
		{name: "nil_interface", val: nil, want: 0},
		{name: "nil_slice", val: nilSlice, want: 0},
		{name: "empty_slice", val: []string{}, want: 0},
		{name: "slice_values", val: []int{1, 2, 3}, want: 3},
		{name: "not_slice", val: 10, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sliceLen(tt.val); got != tt.want {
				t.Fatalf("sliceLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderedTransactionTables(t *testing.T) {
	transactions := map[string]interface{}{
		"tx_special_price_3":        []models.TxSpecialPrice3{{TransactionIDUnique: 10}},
		"tx_bonus_accrual_9":        []models.TxBonusAccrual9{{TransactionIDUnique: 20}},
		"tx_item_registration_1_11": []models.TxItemRegistration1_11{},
		"unknown":                   42,
	}

	got := orderedTransactionTables(transactions)
	want := []string{"tx_bonus_accrual_9", "tx_special_price_3", "unknown"}
	if len(got) != len(want) {
		t.Fatalf("orderedTransactionTables() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("orderedTransactionTables() = %v, want %v", got, want)
		}
	}
}

func TestRetryPolicyBackoffCapsAtMax(t *testing.T) {
	policy := retryPolicy{initialBackoff: 100 * time.Millisecond, maxBackoff: 250 * time.Millisecond}
	if got := policy.retryBackoff(0); got != 100*time.Millisecond {
		t.Fatalf("retryBackoff(0) = %v, want %v", got, 100*time.Millisecond)
	}
	if got := policy.retryBackoff(1); got != 200*time.Millisecond {
		t.Fatalf("retryBackoff(1) = %v, want %v", got, 200*time.Millisecond)
	}
	if got := policy.retryBackoff(3); got != 250*time.Millisecond {
		t.Fatalf("retryBackoff(3) = %v, want %v", got, 250*time.Millisecond)
	}
}

func TestLoadFileDataLoadsTablesInDeterministicOrder(t *testing.T) {
	order := make([]string, 0, 3)
	tx := &fakeTx{}
	loader := newLoaderWithDB(&loaderDBMock{
		beginTxFunc: func(ctx context.Context) (pgx.Tx, error) {
			return tx, nil
		},
		loadTxTableFunc: func(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
			order = append(order, tableName)
			return nil
		},
	})
	loader.policy = retryPolicy{maxRetries: 1, initialBackoff: 0, maxBackoff: 0}

	transactions := map[string]interface{}{
		"tx_special_price_3":        []models.TxSpecialPrice3{{TransactionIDUnique: 10}},
		"tx_bonus_accrual_9":        []models.TxBonusAccrual9{{TransactionIDUnique: 20}},
		"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}},
	}

	if err := loader.LoadFileData(context.Background(), transactions); err != nil {
		t.Fatalf("LoadFileData() unexpected error: %v", err)
	}
	want := []string{"tx_bonus_accrual_9", "tx_item_registration_1_11", "tx_special_price_3"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("LoadFileData() order = %v, want %v", order, want)
		}
	}
	if tx.commitCalls != 1 {
		t.Fatalf("Commit() calls = %d, want 1", tx.commitCalls)
	}
}

func TestLoadFileDataRetriesRetryableErrors(t *testing.T) {
	attempts := 0
	loader := newLoaderWithDB(&loaderDBMock{
		beginTxFunc: func(ctx context.Context) (pgx.Tx, error) {
			return &fakeTx{}, nil
		},
		loadTxTableFunc: func(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
			attempts++
			if attempts == 1 {
				return &pgconn.PgError{Code: "40001"}
			}
			return nil
		},
	})
	loader.policy = retryPolicy{maxRetries: 2, initialBackoff: 0, maxBackoff: 0}

	err := loader.LoadFileData(context.Background(), map[string]interface{}{
		"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}},
	})
	if err != nil {
		t.Fatalf("LoadFileData() unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("LoadFileData() attempts = %d, want 2", attempts)
	}
}

func TestLoadFileDataStopsOnNonRetryableErrors(t *testing.T) {
	attempts := 0
	loader := newLoaderWithDB(&loaderDBMock{
		beginTxFunc: func(ctx context.Context) (pgx.Tx, error) {
			return &fakeTx{}, nil
		},
		loadTxTableFunc: func(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
			attempts++
			return errors.New("constraint violation")
		},
	})
	loader.policy = retryPolicy{maxRetries: 3, initialBackoff: 0, maxBackoff: 0}

	err := loader.LoadFileData(context.Background(), map[string]interface{}{
		"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}},
	})
	if err == nil {
		t.Fatal("LoadFileData() expected error, got nil")
	}
	if attempts != 1 {
		t.Fatalf("LoadFileData() attempts = %d, want 1", attempts)
	}
}

func TestLoadFileDataWithReconcileDeletesStaleRowsDeterministically(t *testing.T) {
	loadOrder := make([]string, 0, 2)
	tx := &fakeTx{}
	loader := newLoaderWithDB(&loaderDBMock{
		beginTxFunc: func(ctx context.Context) (pgx.Tx, error) {
			return tx, nil
		},
		loadTxTableFunc: func(ctx context.Context, tx pgx.Tx, tableName string, data interface{}) error {
			loadOrder = append(loadOrder, tableName)
			return nil
		},
	})
	loader.policy = retryPolicy{maxRetries: 1, initialBackoff: 0, maxBackoff: 0}

	staleManifest := map[string][]int64{
		"tx_special_price_3":        {10},
		"tx_bonus_accrual_9":        {20},
		"tx_item_registration_1_11": {},
	}
	transactions := map[string]interface{}{
		"tx_special_price_3": []models.TxSpecialPrice3{{TransactionIDUnique: 11}},
		"tx_bonus_accrual_9": []models.TxBonusAccrual9{{TransactionIDUnique: 21}},
	}

	if err := loader.LoadFileDataWithReconcile(context.Background(), "P13/P13", staleManifest, transactions); err != nil {
		t.Fatalf("LoadFileDataWithReconcile() unexpected error: %v", err)
	}
	if len(tx.execSQL) != 2 {
		t.Fatalf("Exec() calls = %d, want 2", len(tx.execSQL))
	}
	if !strings.Contains(tx.execSQL[0], "tx_bonus_accrual_9") || !strings.Contains(tx.execSQL[1], "tx_special_price_3") {
		t.Fatalf("Exec() order = %v, want bonus then special", tx.execSQL)
	}
	wantLoadOrder := []string{"tx_bonus_accrual_9", "tx_special_price_3"}
	for i := range wantLoadOrder {
		if loadOrder[i] != wantLoadOrder[i] {
			t.Fatalf("load order = %v, want %v", loadOrder, wantLoadOrder)
		}
	}
}

// TestGetTransactionDetailsEdgeCases tests edge cases for GetTransactionDetails.
func TestGetTransactionDetailsEdgeCases(t *testing.T) {
	loader := &Loader{db: nil}

	tests := []struct {
		name         string
		transactions map[string]interface{}
		expectedLen  int
	}{
		{
			name: "empty slice",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{},
			},
			expectedLen: 0,
		},
		{
			name: "single type with data",
			transactions: map[string]interface{}{
				"tx_item_registration_1_11": []models.TxItemRegistration1_11{{TransactionIDUnique: 1}},
			},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.GetTransactionDetails(tt.transactions)
			if len(result) != tt.expectedLen {
				t.Errorf("GetTransactionDetails() length = %d, want %d", len(result), tt.expectedLen)
			}
		})
	}
}
