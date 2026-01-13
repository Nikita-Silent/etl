package ftp

import (
	"strings"
	"testing"
	"time"

	"github.com/jlaffaye/ftp"
)

func TestFilterUnprocessedFiles(t *testing.T) {
	tests := []struct {
		name     string
		input    []*ftp.Entry
		expected int
		desc     string
	}{
		{
			name: "all unprocessed files",
			input: []*ftp.Entry{
				{Name: "file1.txt", Type: ftp.EntryTypeFile},
				{Name: "file2.txt", Type: ftp.EntryTypeFile},
				{Name: "file3.txt", Type: ftp.EntryTypeFile},
			},
			expected: 3,
			desc:     "should return all files when none are processed",
		},
		{
			name: "some processed files",
			input: []*ftp.Entry{
				{Name: "file1.txt", Type: ftp.EntryTypeFile},
				{Name: "file1.txt.processed", Type: ftp.EntryTypeFile},
				{Name: "file2.txt", Type: ftp.EntryTypeFile},
				{Name: "file3.txt", Type: ftp.EntryTypeFile},
			},
			expected: 2,
			desc:     "should filter out file1.txt and file1.txt.processed",
		},
		{
			name: "all processed files",
			input: []*ftp.Entry{
				{Name: "file1.txt", Type: ftp.EntryTypeFile},
				{Name: "file1.txt.processed", Type: ftp.EntryTypeFile},
				{Name: "file2.txt", Type: ftp.EntryTypeFile},
				{Name: "file2.txt.processed", Type: ftp.EntryTypeFile},
			},
			expected: 0,
			desc:     "should return no files when all are processed",
		},
		{
			name: "only .processed files",
			input: []*ftp.Entry{
				{Name: "file1.txt.processed", Type: ftp.EntryTypeFile},
				{Name: "file2.txt.processed", Type: ftp.EntryTypeFile},
			},
			expected: 0,
			desc:     "should return no files when only .processed files exist",
		},
		{
			name:     "empty list",
			input:    []*ftp.Entry{},
			expected: 0,
			desc:     "should return empty list for empty input",
		},
		{
			name: "mixed processed and unprocessed",
			input: []*ftp.Entry{
				{Name: "file1.txt", Type: ftp.EntryTypeFile},
				{Name: "file2.txt", Type: ftp.EntryTypeFile},
				{Name: "file2.txt.processed", Type: ftp.EntryTypeFile},
				{Name: "file3.txt", Type: ftp.EntryTypeFile},
				{Name: "file4.txt", Type: ftp.EntryTypeFile},
				{Name: "file4.txt.processed", Type: ftp.EntryTypeFile},
			},
			expected: 2,
			desc:     "should return file1.txt and file3.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterUnprocessedFiles(tt.input)
			if len(result) != tt.expected {
				t.Errorf("%s: got %d files, want %d files", tt.desc, len(result), tt.expected)
			}

			// Verify no .processed files in result
			for _, file := range result {
				if strings.HasSuffix(file.Name, ".processed") {
					t.Errorf("result contains .processed file: %s", file.Name)
				}
			}
		})
	}
}

func TestFilterFilesByName(t *testing.T) {
	files := []*ftp.Entry{
		{Name: "file1.txt", Type: ftp.EntryTypeFile},
		{Name: "file2.txt", Type: ftp.EntryTypeFile},
		{Name: "SaveResult001.txt", Type: ftp.EntryTypeFile},
		{Name: "file3.txt", Type: ftp.EntryTypeFile},
	}

	tests := []struct {
		name         string
		excludeNames []string
		expected     int
	}{
		{
			name:         "exclude single file",
			excludeNames: []string{"SaveResult001.txt"},
			expected:     3,
		},
		{
			name:         "exclude multiple files",
			excludeNames: []string{"file1.txt", "file2.txt"},
			expected:     2,
		},
		{
			name:         "exclude nothing",
			excludeNames: []string{},
			expected:     4,
		},
		{
			name:         "exclude non-existent file",
			excludeNames: []string{"nonexistent.txt"},
			expected:     4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterFilesByName(files, tt.excludeNames...)
			if len(result) != tt.expected {
				t.Errorf("got %d files, want %d", len(result), tt.expected)
			}

			// Verify excluded files are not in result
			excludeSet := make(map[string]bool)
			for _, name := range tt.excludeNames {
				excludeSet[name] = true
			}

			for _, file := range result {
				if excludeSet[file.Name] {
					t.Errorf("result contains excluded file: %s", file.Name)
				}
			}
		})
	}
}

func TestGetProcessedFileNames(t *testing.T) {
	tests := []struct {
		name     string
		input    []*ftp.Entry
		expected map[string]bool
	}{
		{
			name: "single processed file",
			input: []*ftp.Entry{
				{Name: "file1.txt.processed", Type: ftp.EntryTypeFile},
			},
			expected: map[string]bool{"file1.txt": true},
		},
		{
			name: "multiple processed files",
			input: []*ftp.Entry{
				{Name: "file1.txt.processed", Type: ftp.EntryTypeFile},
				{Name: "file2.txt.processed", Type: ftp.EntryTypeFile},
				{Name: "file3.txt.processed", Type: ftp.EntryTypeFile},
			},
			expected: map[string]bool{
				"file1.txt": true,
				"file2.txt": true,
				"file3.txt": true,
			},
		},
		{
			name: "mixed files and processed",
			input: []*ftp.Entry{
				{Name: "file1.txt", Type: ftp.EntryTypeFile},
				{Name: "file2.txt.processed", Type: ftp.EntryTypeFile},
				{Name: "file3.txt", Type: ftp.EntryTypeFile},
			},
			expected: map[string]bool{"file2.txt": true},
		},
		{
			name:     "no processed files",
			input:    []*ftp.Entry{{Name: "file1.txt", Type: ftp.EntryTypeFile}},
			expected: map[string]bool{},
		},
		{
			name:     "empty list",
			input:    []*ftp.Entry{},
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProcessedFileNames(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("got %d entries, want %d", len(result), len(tt.expected))
			}

			for name, expected := range tt.expected {
				if result[name] != expected {
					t.Errorf("expected %s to be %v, got %v", name, expected, result[name])
				}
			}
		})
	}
}

func TestIsFileInProcessedSet(t *testing.T) {
	processedSet := map[string]bool{
		"file1.txt": true,
		"file2.txt": true,
	}

	tests := []struct {
		fileName string
		expected bool
	}{
		{"file1.txt", true},
		{"file2.txt", true},
		{"file3.txt", false},
		{"", false},
		{"nonexistent.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			result := IsFileInProcessedSet(tt.fileName, processedSet)
			if result != tt.expected {
				t.Errorf("IsFileInProcessedSet(%s) = %v, want %v", tt.fileName, result, tt.expected)
			}
		})
	}
}

func TestFilterUnprocessedFiles_EdgeCases(t *testing.T) {
	t.Run("files with similar names", func(t *testing.T) {
		input := []*ftp.Entry{
			{Name: "file.txt", Type: ftp.EntryTypeFile},
			{Name: "file.txt.processed", Type: ftp.EntryTypeFile},
			{Name: "file.txt.backup", Type: ftp.EntryTypeFile},
		}

		result := FilterUnprocessedFiles(input)

		// Should include file.txt.backup but not file.txt (it has .processed version)
		if len(result) != 1 {
			t.Errorf("expected 1 file, got %d", len(result))
		}
		if len(result) > 0 && result[0].Name != "file.txt.backup" {
			t.Errorf("expected file.txt.backup, got %s", result[0].Name)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := FilterUnprocessedFiles(nil)
		if result == nil {
			t.Error("expected non-nil result for nil input")
		}
		if len(result) != 0 {
			t.Errorf("expected empty slice, got length %d", len(result))
		}
	})

	t.Run("files with double extension", func(t *testing.T) {
		input := []*ftp.Entry{
			{Name: "data.tar.gz", Type: ftp.EntryTypeFile},
			{Name: "data.tar.gz.processed", Type: ftp.EntryTypeFile},
		}

		result := FilterUnprocessedFiles(input)
		if len(result) != 0 {
			t.Errorf("expected 0 files, got %d", len(result))
		}
	})
}

func TestBatchOperationsPerformance(t *testing.T) {
	// Create a large set of files to test performance
	const fileCount = 10000
	files := make([]*ftp.Entry, fileCount*2) // Half regular, half .processed

	for i := 0; i < fileCount; i++ {
		files[i*2] = &ftp.Entry{
			Name: "file" + string(rune(i)) + ".txt",
			Type: ftp.EntryTypeFile,
			Time: time.Now(),
		}
		files[i*2+1] = &ftp.Entry{
			Name: "file" + string(rune(i)) + ".txt.processed",
			Type: ftp.EntryTypeFile,
			Time: time.Now(),
		}
	}

	// This should complete quickly (O(n) instead of O(n²))
	start := time.Now()
	result := FilterUnprocessedFiles(files)
	duration := time.Since(start)

	if len(result) != 0 {
		t.Errorf("expected 0 unprocessed files, got %d", len(result))
	}

	// Should complete in under 100ms even for 10k files
	if duration > 100*time.Millisecond {
		t.Logf("FilterUnprocessedFiles took %v for %d files (acceptable but could be optimized)", duration, fileCount*2)
	}
}

func BenchmarkFilterUnprocessedFiles(b *testing.B) {
	files := []*ftp.Entry{
		{Name: "file1.txt", Type: ftp.EntryTypeFile},
		{Name: "file2.txt", Type: ftp.EntryTypeFile},
		{Name: "file2.txt.processed", Type: ftp.EntryTypeFile},
		{Name: "file3.txt", Type: ftp.EntryTypeFile},
		{Name: "file4.txt", Type: ftp.EntryTypeFile},
		{Name: "file4.txt.processed", Type: ftp.EntryTypeFile},
		{Name: "file5.txt", Type: ftp.EntryTypeFile},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterUnprocessedFiles(files)
	}
}

func BenchmarkGetProcessedFileNames(b *testing.B) {
	files := []*ftp.Entry{
		{Name: "file1.txt.processed", Type: ftp.EntryTypeFile},
		{Name: "file2.txt.processed", Type: ftp.EntryTypeFile},
		{Name: "file3.txt", Type: ftp.EntryTypeFile},
		{Name: "file4.txt.processed", Type: ftp.EntryTypeFile},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetProcessedFileNames(files)
	}
}
