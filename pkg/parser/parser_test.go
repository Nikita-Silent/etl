package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/go-frontol-loader/pkg/models"
)

func TestParseFileHeader(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDBID    string
		wantReport  string
		wantProcess bool
		wantErr     bool
	}{
		{
			name:        "valid header with processed flag 1",
			content:     "1\nDB123\nREPORT456\n",
			wantDBID:    "DB123",
			wantReport:  "REPORT456",
			wantProcess: true,
			wantErr:     false,
		},
		{
			name:        "valid header with processed flag @",
			content:     "@\nDB789\nREPORT012\n",
			wantDBID:    "DB789",
			wantReport:  "REPORT012",
			wantProcess: true,
			wantErr:     false,
		},
		{
			name:        "valid header not processed",
			content:     "0\nDB456\nREPORT789\n",
			wantDBID:    "DB456",
			wantReport:  "REPORT789",
			wantProcess: false,
			wantErr:     false,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "only one line",
			content: "1\n",
			wantErr: true,
		},
		{
			name:    "only two lines",
			content: "1\nDB123\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.txt")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Open file
			file, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("failed to open temp file: %v", err)
			}
			defer file.Close()

			// Test
			header, err := parseFileHeader(file)

			if tt.wantErr {
				if err == nil {
					t.Error("parseFileHeader() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseFileHeader() unexpected error: %v", err)
				return
			}

			if header.DBID != tt.wantDBID {
				t.Errorf("parseFileHeader() DBID = %v, want %v", header.DBID, tt.wantDBID)
			}
			if header.ReportNum != tt.wantReport {
				t.Errorf("parseFileHeader() ReportNum = %v, want %v", header.ReportNum, tt.wantReport)
			}
			if header.Processed != tt.wantProcess {
				t.Errorf("parseFileHeader() Processed = %v, want %v", header.Processed, tt.wantProcess)
			}
		})
	}
}

func TestParseTransactionLine(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		sourceFolder string
		wantType     int
		wantErr      bool
	}{
		{
			name:         "valid transaction type 1",
			line:         "12345;01.12.2024;10:30:00;1;001;100;1;ITEM001;GRP01;1000.50;5;5025.50;1;10",
			sourceFolder: "test_folder",
			wantType:     1,
			wantErr:      false,
		},
		{
			name:         "valid transaction type 2",
			line:         "12346;01.12.2024;10:31:00;2;001;101;2;PL001;GRP02;1;500.00;450.00;2;1001;2001",
			sourceFolder: "test_folder",
			wantType:     2,
			wantErr:      false,
		},
		{
			name:         "insufficient fields",
			line:         "12345;01.12.2024;10:30:00",
			sourceFolder: "test_folder",
			wantErr:      true,
		},
		{
			name:         "invalid transaction type",
			line:         "12345;01.12.2024;10:30:00;ABC;001;100;1",
			sourceFolder: "test_folder",
			wantErr:      true,
		},
		{
			name:         "empty line",
			line:         "",
			sourceFolder: "test_folder",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTransactionLine(tt.line, tt.sourceFolder)

			if tt.wantErr {
				if err == nil {
					t.Error("parseTransactionLine() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseTransactionLine() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("parseTransactionLine() returned nil result")
			}
		})
	}
}

func TestParseBaseTransactionData(t *testing.T) {
	tests := []struct {
		name         string
		fields       []string
		sourceFolder string
		wantID       string
		wantType     int
		wantErr      bool
	}{
		{
			name:         "valid base data",
			fields:       []string{"12345", "01.12.2024", "10:30:00", "1", "001", "100", "1", "10", "1", "0", "", "0", "EMP001"},
			sourceFolder: "test_folder",
			wantID:       "12345",
			wantType:     1,
			wantErr:      false,
		},
		{
			name:         "minimal fields",
			fields:       []string{"12345", "01.12.2024", "10:30:00", "1"},
			sourceFolder: "test_folder",
			wantID:       "12345",
			wantType:     1,
			wantErr:      false,
		},
		{
			name:         "insufficient fields",
			fields:       []string{"12345", "01.12.2024", "10:30:00"},
			sourceFolder: "test_folder",
			wantErr:      true,
		},
		{
			name:         "invalid transaction type",
			fields:       []string{"12345", "01.12.2024", "10:30:00", "ABC"},
			sourceFolder: "test_folder",
			wantErr:      true,
		},
		{
			name:         "empty date uses current",
			fields:       []string{"12345", "", "", "1", "001"},
			sourceFolder: "test_folder",
			wantID:       "12345",
			wantType:     1,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBaseTransactionData(tt.fields, tt.sourceFolder)

			if tt.wantErr {
				if err == nil {
					t.Error("parseBaseTransactionData() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseBaseTransactionData() unexpected error: %v", err)
				return
			}

			if result.ID != tt.wantID {
				t.Errorf("parseBaseTransactionData() ID = %v, want %v", result.ID, tt.wantID)
			}
			if result.TransactionType != tt.wantType {
				t.Errorf("parseBaseTransactionData() TransactionType = %v, want %v", result.TransactionType, tt.wantType)
			}
			if result.SourceFolder != tt.sourceFolder {
				t.Errorf("parseBaseTransactionData() SourceFolder = %v, want %v", result.SourceFolder, tt.sourceFolder)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	// Create a complete test file
	content := `1
DB_TEST_123
REPORT_001
12345;01.12.2024;10:30:00;1;001;100;1;ITEM001;GRP01;1000.50;5;5025.50;1;10;100.10;500.50;1;SKU001;1234567890;1000.00;01;0;0;0;;info;1;EMP001;0;;0;0;;;0;0;0;;;0;;;;
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_transactions.txt")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	transactions, header, err := ParseFile(tmpFile, "test_folder")

	if err != nil {
		t.Fatalf("ParseFile() unexpected error: %v", err)
	}

	if header == nil {
		t.Fatal("ParseFile() returned nil header")
	}

	if header.DBID != "DB_TEST_123" {
		t.Errorf("ParseFile() header.DBID = %v, want DB_TEST_123", header.DBID)
	}

	if header.ReportNum != "REPORT_001" {
		t.Errorf("ParseFile() header.ReportNum = %v, want REPORT_001", header.ReportNum)
	}

	if transactions == nil {
		t.Fatal("ParseFile() returned nil transactions")
	}

	// Check that we got tx_item_registration_1_11
	if _, ok := transactions["tx_item_registration_1_11"]; !ok {
		t.Error("ParseFile() expected tx_item_registration_1_11 in result")
	}
}

// Benchmarks
func BenchmarkParseTransactionLine(b *testing.B) {
	line := "12345;01.12.2024;10:30:00;1;001;100;1;ITEM001;GRP01;1000.50;5;5025.50;1;10;100.10;500.50;1;SKU001;1234567890;1000.00;01;0;0;0;;info;1;EMP001;0;;0;0;;;0;0;0;;;0;;;;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseTransactionLine(line, "test_folder")
	}
}

func BenchmarkParseBaseTransactionData(b *testing.B) {
	fields := []string{"12345", "01.12.2024", "10:30:00", "1", "001", "100", "1", "10", "1", "0", "", "0", "EMP001"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseBaseTransactionData(fields, "test_folder")
	}
}

func TestParseTransactionLineTxTable(t *testing.T) {
	line := "12345;01.12.2024;10:30:00;1;001;100;1;ITEM001;GRP01;1000.50;5;5025.50;1;10;100.10;500.50;1;SKU001;1234567890;1000.00;01;0;0;0;;info;1;EMP001;0;;0;0;;;0;0;0;;;0;;;;"

	parsed, err := ParseTransactionLine(line, "test_folder")
	if err != nil {
		t.Fatalf("ParseTransactionLine() unexpected error: %v", err)
	}

	tx, ok := parsed.(ParsedTransaction)
	if !ok {
		t.Fatalf("ParseTransactionLine() returned unexpected type %T", parsed)
	}

	if tx.Table != "tx_item_registration_1_11" {
		t.Fatalf("ParseTransactionLine() table = %s, want tx_item_registration_1_11", tx.Table)
	}

	if _, ok := tx.Value.(models.TxItemRegistration1_11); !ok {
		t.Fatalf("ParseTransactionLine() value type = %T, want models.TxItemRegistration1_11", tx.Value)
	}
}
