package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/user/go-frontol-loader/pkg/repository"
)

type recordingFlusher struct {
	flushCalls int
}

func (f *recordingFlusher) Flush() {
	f.flushCalls++
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestStreamExportTXT_PreservesFrontolFormat(t *testing.T) {
	var buf bytes.Buffer
	flusher := &recordingFlusher{}
	transactions := []repository.ExportRow{
		{TransactionIDUnique: 1, RawLine: "1;01.12.2024;10:00:00;42"},
		{TransactionIDUnique: 2, RawLine: "2;01.12.2024;10:01:00;49"},
	}

	stats, err := streamExportTXT(&buf, flusher, transactions)
	if err != nil {
		t.Fatalf("streamExportTXT() unexpected error: %v", err)
	}
	if stats.Written != 2 || stats.Skipped != 0 {
		t.Fatalf("streamExportTXT() stats = %+v, want written=2 skipped=0", stats)
	}

	want := strings.Join([]string{
		"#",
		"1",
		"2",
		"1;01.12.2024;10:00:00;42",
		"2;01.12.2024;10:01:00;49",
		"",
	}, "\n")
	if got := buf.String(); got != want {
		t.Fatalf("streamExportTXT() output = %q, want %q", got, want)
	}
	if flusher.flushCalls != 1 {
		t.Fatalf("flusher calls = %d, want %d", flusher.flushCalls, 1)
	}
}

func TestStreamExportTXT_TracksSkippedRows(t *testing.T) {
	var buf bytes.Buffer
	transactions := []repository.ExportRow{
		{TransactionIDUnique: 1, RawLine: "1;01.12.2024;10:00:00;42"},
		{TransactionIDUnique: 2, RawLine: ""},
		{TransactionIDUnique: 3, RawLine: "3;01.12.2024;10:02:00;49"},
	}

	stats, err := streamExportTXT(&buf, nil, transactions)
	if err != nil {
		t.Fatalf("streamExportTXT() unexpected error: %v", err)
	}
	if stats.Written != 2 || stats.Skipped != 1 {
		t.Fatalf("streamExportTXT() stats = %+v, want written=2 skipped=1", stats)
	}
	if len(stats.SkippedIDSamples) != 1 || stats.SkippedIDSamples[0] != 2 {
		t.Fatalf("streamExportTXT() skipped ids = %v, want [2]", stats.SkippedIDSamples)
	}
	if !strings.Contains(buf.String(), "\n3\n") {
		t.Fatalf("streamExportTXT() header count not preserved: %q", buf.String())
	}
}

func TestStreamExportTXT_WriteError(t *testing.T) {
	_, err := streamExportTXT(failingWriter{}, nil, []repository.ExportRow{{TransactionIDUnique: 1, RawLine: "1;line"}})
	if err == nil {
		t.Fatal("streamExportTXT() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "flush export writer") {
		t.Fatalf("streamExportTXT() error = %v, want flush write error", err)
	}
}

func TestStreamExportTXT_FlushesPeriodically(t *testing.T) {
	var buf bytes.Buffer
	flusher := &recordingFlusher{}
	transactions := make([]repository.ExportRow, exportFlushInterval+1)
	for i := range transactions {
		transactions[i] = repository.ExportRow{
			TransactionIDUnique: int64(i + 1),
			TransactionTime:     time.Unix(int64(i), 0),
			RawLine:             "line",
		}
	}

	_, err := streamExportTXT(&buf, flusher, transactions)
	if err != nil {
		t.Fatalf("streamExportTXT() unexpected error: %v", err)
	}
	if flusher.flushCalls < 2 {
		t.Fatalf("flusher calls = %d, want at least 2", flusher.flushCalls)
	}
}
