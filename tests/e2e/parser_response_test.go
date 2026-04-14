package e2e

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/user/go-frontol-loader/pkg/parser"
)

func TestE2E_ParseResponseFile(t *testing.T) {
	root := findRepoRoot(t)
	filePath := filepath.Join(root, "data", "response.txt")

	transactions, header, err := parser.ParseFile(filePath, "P13")
	if err != nil {
		t.Fatalf("ParseFile() unexpected error: %v", err)
	}
	if header == nil {
		t.Fatal("ParseFile() returned nil header")
	}
	if header.Processed {
		t.Fatalf("ParseFile() header.Processed = true, want false for sample file")
	}
	if header.DBID != "1" {
		t.Fatalf("ParseFile() header.DBID = %q, want %q", header.DBID, "1")
	}
	if header.ReportNum != "24335" {
		t.Fatalf("ParseFile() header.ReportNum = %q, want %q", header.ReportNum, "24335")
	}
	if len(transactions) == 0 {
		t.Fatalf("ParseFile() returned no transactions")
	}

	if total := transactionCount(transactions); total != 1346 {
		t.Fatalf("ParseFile() total transactions = %d, want %d", total, 1346)
	}

	expectedCounts := map[string]int{
		"tx_shift_open_doc_64":       3,
		"tx_document_open_42":        82,
		"tx_bill_registration_21_23": 2,
		"tx_document_close_gp_49":    82,
		"tx_cash_in_50":              2,
		"tx_item_registration_1_11":  144,
	}

	for table, want := range expectedCounts {
		got := sliceCount(transactions[table])
		if got != want {
			t.Fatalf("ParseFile() table %s count = %d, want %d", table, got, want)
		}
	}
}

func transactionCount(transactions map[string]interface{}) int {
	total := 0
	for _, items := range transactions {
		total += sliceCount(items)
	}
	return total
}

func sliceCount(items interface{}) int {
	if items == nil {
		return 0
	}
	value := reflect.ValueOf(items)
	if value.Kind() != reflect.Slice {
		return 0
	}
	return value.Len()
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "data", "response.txt")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatalf("repo root not found from %s", wd)
	return ""
}
