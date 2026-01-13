package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/user/go-frontol-loader/pkg/models"
)

func TestConvertToUTF8(t *testing.T) {
	cp1251 := string([]byte{0xCF, 0xF0, 0xE8, 0xE2, 0xE5, 0xF2})

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "ascii", input: "hello", want: "hello"},
		{name: "cp1251_bytes", input: cp1251, want: "Привет"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToUTF8(tt.input); got != tt.want {
				t.Fatalf("convertToUTF8() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("utf8_string_is_corrupted", func(t *testing.T) {
		input := "Привет"
		got := convertToUTF8(input)
		if got == input {
			t.Fatalf("convertToUTF8() should corrupt valid UTF-8, got %q", got)
		}
	})
}

func TestSafeValue(t *testing.T) {
	cp1251 := string([]byte{0xCF, 0xF0, 0xE8, 0xE2, 0xE5, 0xF2})

	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{name: "empty_string", input: "", want: nil},
		{name: "ascii_string", input: "hello", want: "hello"},
		{name: "cp1251_string", input: cp1251, want: "Привет"},
		{name: "float_zero", input: 0.0, want: nil},
		{name: "float_value", input: 1.5, want: 1.5},
		{name: "int_zero", input: 0, want: nil},
		{name: "int_value", input: 7, want: 7},
		{name: "int64_zero", input: int64(0), want: nil},
		{name: "int64_value", input: int64(9), want: int64(9)},
		{name: "passthrough", input: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeValue(tt.input); got != tt.want {
				t.Fatalf("safeValue() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSafeValueAllowZero(t *testing.T) {
	cp1251 := string([]byte{0xCF, 0xF0, 0xE8, 0xE2, 0xE5, 0xF2})

	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{name: "empty_string", input: "", want: nil},
		{name: "ascii_string", input: "hello", want: "hello"},
		{name: "cp1251_string", input: cp1251, want: "Привет"},
		{name: "float_zero", input: 0.0, want: 0.0},
		{name: "float_value", input: 2.5, want: 2.5},
		{name: "int_zero", input: 0, want: 0},
		{name: "int_value", input: 4, want: 4},
		{name: "int64_zero", input: int64(0), want: int64(0)},
		{name: "int64_value", input: int64(11), want: int64(11)},
		{name: "passthrough", input: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeValueAllowZero(tt.input); got != tt.want {
				t.Fatalf("safeValueAllowZero() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

// TestPoolInterface tests that Pool implements DatabasePool interface.
func TestPoolInterface(t *testing.T) {
	var _ DatabasePool = (*Pool)(nil)
}

// TestMockPool tests MockPool implementation.
func TestMockPool(t *testing.T) {
	mock := &MockPool{}

	// Test that mock implements interface.
	var _ DatabasePool = mock

	// Test Close doesn't panic.
	mock.Close()

	// Test BeginTx with nil function.
	_, err := mock.BeginTx(context.Background())
	if err != nil {
		t.Errorf("MockPool.BeginTx() unexpected error: %v", err)
	}

	// Test BeginTx with custom function.
	mock.BeginTxFunc = func(ctx context.Context) (pgx.Tx, error) {
		return nil, nil
	}
	_, err = mock.BeginTx(context.Background())
	if err != nil {
		t.Errorf("MockPool.BeginTx() with function unexpected error: %v", err)
	}
}

// TestLoadDataWithMock tests LoadData with mock.
func TestLoadDataWithMock(t *testing.T) {
	mock := &MockPool{}
	called := false

	mock.LoadDataFunc = func(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error {
		called = true
		if tableName != "test_table" {
			t.Errorf("Expected tableName 'test_table', got '%s'", tableName)
		}
		if len(columns) != 2 {
			t.Errorf("Expected 2 columns, got %d", len(columns))
		}
		if len(rows) != 1 {
			t.Errorf("Expected 1 row, got %d", len(rows))
		}
		return nil
	}

	err := mock.LoadData(context.Background(), nil, "test_table", []string{"col1", "col2"}, [][]interface{}{{"val1", "val2"}})
	if err != nil {
		t.Errorf("MockPool.LoadData() unexpected error: %v", err)
	}
	if !called {
		t.Error("MockPool.LoadData() function was not called")
	}
}

// TestLoadDataEmptyRows tests LoadData with empty rows.
func TestLoadDataEmptyRows(t *testing.T) {
	mock := &MockPool{}

	mock.LoadDataFunc = func(ctx context.Context, tx pgx.Tx, tableName string, columns []string, rows [][]interface{}) error {
		return nil
	}

	// Empty rows should not call the function (based on real implementation).
	err := mock.LoadData(context.Background(), nil, "test_table", []string{"col1"}, [][]interface{}{})
	if err != nil {
		t.Errorf("MockPool.LoadData() with empty rows unexpected error: %v", err)
	}
}

func TestBuildTxRowAllowZero(t *testing.T) {
	findIndex := func(schema []models.TxColumnSpec, name string) int {
		for i, spec := range schema {
			if spec.Name == name {
				return i
			}
		}
		return -1
	}

	fiscalSchema := models.TxSchemas["tx_fiscal_payment_40"]
	fiscalRow, err := buildTxRow(fiscalSchema, models.TxFiscalPayment40{
		OperationType:        0,
		PaymentTypeOperation: 0,
		CashOutAmount:        0,
	})
	if err != nil {
		t.Fatalf("buildTxRow fiscal_payment_40 error: %v", err)
	}

	for _, col := range []string{"operation_type", "payment_type_operation", "cash_out_amount"} {
		idx := findIndex(fiscalSchema, col)
		if idx < 0 {
			t.Fatalf("column %s not found in tx_fiscal_payment_40 schema", col)
		}
		if fiscalRow[idx] == nil {
			t.Fatalf("column %s should preserve zero, got nil", col)
		}
	}

	itemSchema := models.TxSchemas["tx_item_registration_1_11"]
	itemRow, err := buildTxRow(itemSchema, models.TxItemRegistration1_11{
		ItemTypeCode: 0,
	})
	if err != nil {
		t.Fatalf("buildTxRow item_registration_1_11 error: %v", err)
	}
	itemIdx := findIndex(itemSchema, "item_type_code")
	if itemIdx < 0 {
		t.Fatalf("column item_type_code not found in tx_item_registration_1_11 schema")
	}
	if itemRow[itemIdx] == nil {
		t.Fatal("column item_type_code should preserve zero, got nil")
	}
}
