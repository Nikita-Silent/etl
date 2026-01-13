package repository

import (
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/user/go-frontol-loader/pkg/models"
)

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
