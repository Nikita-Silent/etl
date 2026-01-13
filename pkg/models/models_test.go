package models

import (
	"testing"
	"time"
)

func TestProcessingStats_AvgTimePerFile(t *testing.T) {
	tests := []struct {
		name           string
		filesProcessed int
		startOffset    time.Duration
		wantZero       bool
	}{
		{
			name:           "zero files returns zero duration",
			filesProcessed: 0,
			startOffset:    -1 * time.Second,
			wantZero:       true,
		},
		{
			name:           "one file processed",
			filesProcessed: 1,
			startOffset:    -1 * time.Second,
			wantZero:       false,
		},
		{
			name:           "multiple files processed",
			filesProcessed: 10,
			startOffset:    -10 * time.Second,
			wantZero:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &ProcessingStats{
				StartTime:      time.Now().Add(tt.startOffset),
				FilesProcessed: tt.filesProcessed,
			}

			got := ps.AvgTimePerFile()

			if tt.wantZero && got != 0 {
				t.Errorf("AvgTimePerFile() = %v, want 0", got)
			}
			if !tt.wantZero && got == 0 {
				t.Error("AvgTimePerFile() = 0, want non-zero")
			}
		})
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := Config{
		DBHost:      "localhost",
		DBPort:      5432,
		DBUser:      "user",
		DBPassword:  "password",
		DBName:      "testdb",
		DBSSLMode:   "disable",
		FTPHost:     "ftphost",
		FTPPort:     21,
		FTPUser:     "ftpuser",
		FTPPassword: "ftppass",
	}

	// Verify all fields are set correctly
	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %v, want localhost", cfg.DBHost)
	}
	if cfg.DBPort != 5432 {
		t.Errorf("DBPort = %v, want 5432", cfg.DBPort)
	}
	if cfg.FTPPort != 21 {
		t.Errorf("FTPPort = %v, want 21", cfg.FTPPort)
	}
}

func TestKassaFolder_Fields(t *testing.T) {
	folder := KassaFolder{
		KassaCode:    "001",
		FolderName:   "folder1",
		RequestPath:  "/request/001/folder1",
		ResponsePath: "/response/001/folder1",
	}

	if folder.KassaCode != "001" {
		t.Errorf("KassaCode = %v, want 001", folder.KassaCode)
	}
	if folder.RequestPath != "/request/001/folder1" {
		t.Errorf("RequestPath = %v, want /request/001/folder1", folder.RequestPath)
	}
}

func TestFileHeader_Fields(t *testing.T) {
	header := FileHeader{
		Processed: true,
		DBID:      "DB123",
		ReportNum: "REP456",
	}

	if !header.Processed {
		t.Error("Processed should be true")
	}
	if header.DBID != "DB123" {
		t.Errorf("DBID = %v, want DB123", header.DBID)
	}
	if header.ReportNum != "REP456" {
		t.Errorf("ReportNum = %v, want REP456", header.ReportNum)
	}
}

func TestColumnToFieldName(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected string
	}{
		{name: "simple", column: "transaction_id_unique", expected: "TransactionIDUnique"},
		{name: "kkt", column: "kkt_section", expected: "KKTSection"},
		{name: "vat", column: "vat_20_amount", expected: "VAT20Amount"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColumnToFieldName(tt.column); got != tt.expected {
				t.Errorf("ColumnToFieldName(%q) = %q, want %q", tt.column, got, tt.expected)
			}
		})
	}
}

// Benchmarks
func BenchmarkProcessingStats_AvgTimePerFile(b *testing.B) {
	ps := &ProcessingStats{
		StartTime:      time.Now().Add(-10 * time.Second),
		FilesProcessed: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ps.AvgTimePerFile()
	}
}
