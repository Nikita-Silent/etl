package parser

import (
	"strings"
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
)

type sampleTx struct {
	Name         string
	Count        int64
	Amount       float64
	Date         time.Time
	Time         time.Time
	SourceFolder string
}

type badIntTx struct {
	Count string
}

type badSourceTx struct {
	SourceFolder int
}

func TestFillTxStructSuccess(t *testing.T) {
	schema := []models.TxColumnSpec{
		{Name: "name", Kind: models.TxColumnString},
		{Name: "count", Kind: models.TxColumnInt64},
		{Name: "amount", Kind: models.TxColumnFloat64},
		{Name: "date", Kind: models.TxColumnDate},
		{Name: "time", Kind: models.TxColumnTime},
		{Name: "source_folder", Kind: models.TxColumnSource},
	}
	fields := []string{"Alice", "10", "2,5", "15.03.2024", "12:34:56"}

	var dest sampleTx
	if err := fillTxStruct(&dest, fields, "folder1", schema); err != nil {
		t.Fatalf("fillTxStruct() unexpected error: %v", err)
	}

	if dest.Name != "Alice" {
		t.Fatalf("Name = %q, want %q", dest.Name, "Alice")
	}
	if dest.Count != 10 {
		t.Fatalf("Count = %d, want 10", dest.Count)
	}
	if dest.Amount != 2.5 {
		t.Fatalf("Amount = %v, want 2.5", dest.Amount)
	}
	wantDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	if !dest.Date.Equal(wantDate) {
		t.Fatalf("Date = %v, want %v", dest.Date, wantDate)
	}
	wantTime := time.Date(0, 1, 1, 12, 34, 56, 0, time.UTC)
	if dest.Time.Hour() != wantTime.Hour() || dest.Time.Minute() != wantTime.Minute() || dest.Time.Second() != wantTime.Second() {
		t.Fatalf("Time = %v, want 12:34:56", dest.Time)
	}
	if dest.SourceFolder != "folder1" {
		t.Fatalf("SourceFolder = %q, want %q", dest.SourceFolder, "folder1")
	}
}

func TestFillTxStructInvalidData(t *testing.T) {
	schema := []models.TxColumnSpec{
		{Name: "count", Kind: models.TxColumnInt64},
		{Name: "amount", Kind: models.TxColumnFloat64},
		{Name: "date", Kind: models.TxColumnDate},
		{Name: "time", Kind: models.TxColumnTime},
	}
	fields := []string{"9223372036854775808", "bad", "32.13.2024", "25:00:00"}

	var dest sampleTx
	if err := fillTxStruct(&dest, fields, "", schema); err != nil {
		t.Fatalf("fillTxStruct() unexpected error: %v", err)
	}

	if dest.Count != 0 {
		t.Fatalf("Count = %d, want 0 on overflow", dest.Count)
	}
	if dest.Amount != 0 {
		t.Fatalf("Amount = %v, want 0 on invalid float", dest.Amount)
	}
	if !dest.Date.IsZero() {
		t.Fatalf("Date = %v, want zero time", dest.Date)
	}
	if !dest.Time.IsZero() {
		t.Fatalf("Time = %v, want zero time", dest.Time)
	}
}

func TestFillTxStructFieldCountEdges(t *testing.T) {
	schema := []models.TxColumnSpec{
		{Name: "name", Kind: models.TxColumnString},
		{Name: "count", Kind: models.TxColumnInt64},
	}

	t.Run("fields_shorter_than_schema", func(t *testing.T) {
		var dest sampleTx
		if err := fillTxStruct(&dest, []string{"Bob"}, "", schema); err != nil {
			t.Fatalf("fillTxStruct() unexpected error: %v", err)
		}
		if dest.Name != "Bob" || dest.Count != 0 {
			t.Fatalf("got Name=%q Count=%d, want Name=Bob Count=0", dest.Name, dest.Count)
		}
	})

	t.Run("fields_longer_than_schema", func(t *testing.T) {
		var dest sampleTx
		if err := fillTxStruct(&dest, []string{"Bob", "3", "extra"}, "", schema); err != nil {
			t.Fatalf("fillTxStruct() unexpected error: %v", err)
		}
		if dest.Name != "Bob" || dest.Count != 3 {
			t.Fatalf("got Name=%q Count=%d, want Name=Bob Count=3", dest.Name, dest.Count)
		}
	})
}

func TestFillTxStructNilSchema(t *testing.T) {
	var dest sampleTx
	dest.Name = "keep"
	if err := fillTxStruct(&dest, []string{"ignored"}, "", nil); err != nil {
		t.Fatalf("fillTxStruct() unexpected error: %v", err)
	}
	if dest.Name != "keep" {
		t.Fatalf("Name = %q, want %q", dest.Name, "keep")
	}
}

func TestFillTxStructErrors(t *testing.T) {
	tests := []struct {
		name    string
		dst     interface{}
		schema  []models.TxColumnSpec
		fields  []string
		wantSub string
	}{
		{
			name:    "not_pointer",
			dst:     sampleTx{},
			schema:  []models.TxColumnSpec{{Name: "name", Kind: models.TxColumnString}},
			fields:  []string{"a"},
			wantSub: "pointer to struct",
		},
		{
			name:    "not_struct",
			dst:     &[]string{},
			schema:  []models.TxColumnSpec{{Name: "name", Kind: models.TxColumnString}},
			fields:  []string{"a"},
			wantSub: "pointer to struct",
		},
		{
			name:    "missing_field",
			dst:     &sampleTx{},
			schema:  []models.TxColumnSpec{{Name: "missing_field", Kind: models.TxColumnString}},
			fields:  []string{"a"},
			wantSub: "missing field",
		},
		{
			name:    "wrong_type_int64",
			dst:     &badIntTx{},
			schema:  []models.TxColumnSpec{{Name: "count", Kind: models.TxColumnInt64}},
			fields:  []string{"1"},
			wantSub: "not int64",
		},
		{
			name:    "source_not_string",
			dst:     &badSourceTx{},
			schema:  []models.TxColumnSpec{{Name: "source_folder", Kind: models.TxColumnSource}},
			fields:  []string{},
			wantSub: "source_folder field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fillTxStruct(tt.dst, tt.fields, "folder", tt.schema)
			if err == nil {
				t.Fatalf("fillTxStruct() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("fillTxStruct() error = %q, want substring %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestFillTxStructNilPointerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil pointer, got none")
		}
	}()

	var dest *sampleTx
	_ = fillTxStruct(dest, []string{"a"}, "", []models.TxColumnSpec{{Name: "name", Kind: models.TxColumnString}})
}

func TestParseTxModel(t *testing.T) {
	fields := []string{"100", "01.01.2024", "10:00:00", "1", "5"}
	result, err := parseTxModel[models.TxItemRegistration1_11](fields, "folderA", "tx_item_registration_1_11")
	if err != nil {
		t.Fatalf("parseTxModel() unexpected error: %v", err)
	}
	if result.TransactionIDUnique != 100 {
		t.Fatalf("TransactionIDUnique = %d, want 100", result.TransactionIDUnique)
	}
	if result.SourceFolder != "folderA" {
		t.Fatalf("SourceFolder = %q, want %q", result.SourceFolder, "folderA")
	}
	if result.TransactionType != 1 {
		t.Fatalf("TransactionType = %d, want 1", result.TransactionType)
	}

	_, err = parseTxModel[models.TxItemRegistration1_11](fields, "folderA", "unknown_table")
	if err == nil || !strings.Contains(err.Error(), "unknown tx schema") {
		t.Fatalf("parseTxModel() expected unknown schema error, got %v", err)
	}
}

func TestWrapParsed(t *testing.T) {
	fields := []string{"100", "01.01.2024", "10:00:00", "1"}
	result, err := wrapParsed[models.TxItemRegistration1_11](fields, "folderA", "tx_item_registration_1_11")
	if err != nil {
		t.Fatalf("wrapParsed() unexpected error: %v", err)
	}
	if result.Table != "tx_item_registration_1_11" {
		t.Fatalf("Table = %q, want %q", result.Table, "tx_item_registration_1_11")
	}
	if _, ok := result.Value.(models.TxItemRegistration1_11); !ok {
		t.Fatalf("Value type = %T, want TxItemRegistration1_11", result.Value)
	}
}

func TestAppendTx(t *testing.T) {
	var dst []models.TxItemRegistration1_11
	value := models.TxItemRegistration1_11{TransactionIDUnique: 1}
	if err := appendTx("tx_item_registration_1_11", value, &dst); err != nil {
		t.Fatalf("appendTx() unexpected error: %v", err)
	}
	if len(dst) != 1 || dst[0].TransactionIDUnique != 1 {
		t.Fatalf("appendTx() appended value mismatch: %#v", dst)
	}

	err := appendTx("tx_item_registration_1_11", models.TxItemTax4_14{}, &dst)
	if err == nil || !strings.Contains(err.Error(), "invalid value") {
		t.Fatalf("appendTx() expected type error, got %v", err)
	}
}
