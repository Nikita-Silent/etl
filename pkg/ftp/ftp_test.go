package ftp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ftplib "github.com/jlaffaye/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

func TestCreateRequestFile(t *testing.T) {
	tests := []struct {
		name        string
		date        string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "valid date",
			date:        "2024-12-01",
			wantContent: "$$$TRANSACTIONSBYDATERANGE\n01.12.2024; 01.12.2024",
			wantErr:     false,
		},
		{
			name:        "future date",
			date:        "2030-01-15",
			wantContent: "$$$TRANSACTIONSBYDATERANGE\n15.01.2030; 15.01.2030",
			wantErr:     false,
		},
		{
			name:        "leap year date",
			date:        "2024-02-29",
			wantContent: "$$$TRANSACTIONSBYDATERANGE\n29.02.2024; 29.02.2024",
			wantErr:     false,
		},
		{
			name:        "empty date uses current",
			date:        "",
			wantContent: "$$$TRANSACTIONSBYDATERANGE\n" + time.Now().Format("02.01.2006") + "; " + time.Now().Format("02.01.2006"),
			wantErr:     false,
		},
		{
			name:    "invalid date format",
			date:    "01-12-2024",
			wantErr: true,
		},
		{
			name:    "invalid date format with slashes",
			date:    "2024/12/01",
			wantErr: true,
		},
		{
			name:    "invalid date value",
			date:    "2024-13-01",
			wantErr: true,
		},
		{
			name:    "non_leap_year_date",
			date:    "2023-02-29",
			wantErr: true,
		},
		{
			name:    "invalid date",
			date:    "not-a-date",
			wantErr: true,
		},
		{
			name:    "garbage",
			date:    "garbage",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			path, err := CreateRequestFile(tmpDir, tt.date)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateRequestFile() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateRequestFile() unexpected error: %v", err)
				return
			}

			// Check file was created
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Error("CreateRequestFile() file was not created")
				return
			}

			// Check content
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("Failed to read created file: %v", err)
				return
			}

			if string(content) != tt.wantContent {
				t.Errorf("CreateRequestFile() content = %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}

// TestFTPClientInterface tests that Client implements FTPClient interface
func TestFTPClientInterface(t *testing.T) {
	var _ FTPClient = (*Client)(nil)
}

// TestMockClient tests MockClient implementation
func TestMockClient(t *testing.T) {
	mock := &MockClient{}

	// Test that mock implements interface
	var _ FTPClient = mock

	// Test Close
	err := mock.Close()
	if err != nil {
		t.Errorf("MockClient.Close() unexpected error: %v", err)
	}

	// Test Close with custom function
	mock.CloseFunc = func() error {
		return nil
	}
	err = mock.Close()
	if err != nil {
		t.Errorf("MockClient.Close() with function unexpected error: %v", err)
	}
}

// TestMockClientListFiles tests ListFiles with mock
func TestMockClientListFiles(t *testing.T) {
	mock := &MockClient{}
	called := false

	mock.ListFilesFunc = func(path string) ([]*ftplib.Entry, error) {
		called = true
		if path != "/test/path" {
			t.Errorf("Expected path '/test/path', got '%s'", path)
		}
		return []*ftplib.Entry{}, nil
	}

	files, err := mock.ListFiles("/test/path")
	if err != nil {
		t.Errorf("MockClient.ListFiles() unexpected error: %v", err)
	}
	if !called {
		t.Error("MockClient.ListFiles() function was not called")
	}
	if files == nil {
		t.Error("MockClient.ListFiles() returned nil")
	}
}

// TestMockClientDownloadFile tests DownloadFile with mock
func TestMockClientDownloadFile(t *testing.T) {
	mock := &MockClient{}
	called := false

	mock.DownloadFileFunc = func(remotePath, localPath string) error {
		called = true
		if remotePath != "/remote/file.txt" {
			t.Errorf("Expected remotePath '/remote/file.txt', got '%s'", remotePath)
		}
		if localPath != "/local/file.txt" {
			t.Errorf("Expected localPath '/local/file.txt', got '%s'", localPath)
		}
		return nil
	}

	err := mock.DownloadFile("/remote/file.txt", "/local/file.txt")
	if err != nil {
		t.Errorf("MockClient.DownloadFile() unexpected error: %v", err)
	}
	if !called {
		t.Error("MockClient.DownloadFile() function was not called")
	}
}

// TestMockClientMarkFileAsProcessed tests MarkFileAsProcessed with mock
func TestMockClientMarkFileAsProcessed(t *testing.T) {
	mock := &MockClient{}
	called := false

	mock.MarkFileAsProcessedFunc = func(remotePath string) error {
		called = true
		return nil
	}

	err := mock.MarkFileAsProcessed("/test/file.txt")
	if err != nil {
		t.Errorf("MockClient.MarkFileAsProcessed() unexpected error: %v", err)
	}
	if !called {
		t.Error("MockClient.MarkFileAsProcessed() function was not called")
	}
}

// TestMockClientIsFileProcessed tests IsFileProcessed with mock
func TestMockClientIsFileProcessed(t *testing.T) {
	mock := &MockClient{}

	mock.IsFileProcessedFunc = func(remotePath string) bool {
		return true
	}

	result := mock.IsFileProcessed("/test/file.txt")
	if !result {
		t.Error("MockClient.IsFileProcessed() expected true, got false")
	}
}

// TestMockClientAllMethods tests all mock methods don't panic
func TestMockClientAllMethods(t *testing.T) {
	mock := &MockClient{}

	// Test all methods don't panic
	_ = mock.DeleteProcessedFiles("/test")
	_ = mock.ClearAllKassaResponseProcessedFiles()
	_ = mock.UploadFile("/local", "/remote")
	_ = mock.SendRequestToKassa(models.KassaFolder{}, "2024-12-01")
	_ = mock.ClearDirectory("/test")
	_ = mock.ClearAllKassaRequestFolders()
	_ = mock.ClearAllKassaResponseFolders()
	_ = mock.ClearAllKassaFolders()
	_ = mock.SendRequestsToAllKassas()
	_ = mock.SendRequestsToAllKassasWithDate("2024-12-01")
	_ = mock.EnsureDirectoryExists("/test")
	_ = mock.EnsureKassaFoldersExist()
}

func TestGetAllKassaFolders(t *testing.T) {
	tests := []struct {
		name           string
		kassaStructure map[string][]string
		requestDir     string
		responseDir    string
		wantLen        int
	}{
		{
			name: "single kassa single folder",
			kassaStructure: map[string][]string{
				"001": {"folder1"},
			},
			requestDir:  "/request",
			responseDir: "/response",
			wantLen:     1,
		},
		{
			name: "single kassa multiple folders",
			kassaStructure: map[string][]string{
				"001": {"folder1", "folder2", "folder3"},
			},
			requestDir:  "/request",
			responseDir: "/response",
			wantLen:     3,
		},
		{
			name: "multiple kassas",
			kassaStructure: map[string][]string{
				"001": {"folder1", "folder2"},
				"002": {"folder3"},
			},
			requestDir:  "/request",
			responseDir: "/response",
			wantLen:     3,
		},
		{
			name:           "empty structure",
			kassaStructure: map[string][]string{},
			requestDir:     "/request",
			responseDir:    "/response",
			wantLen:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &models.Config{
				KassaStructure: tt.kassaStructure,
				FTPRequestDir:  tt.requestDir,
				FTPResponseDir: tt.responseDir,
			}

			got := GetAllKassaFolders(cfg)

			if len(got) != tt.wantLen {
				t.Errorf("GetAllKassaFolders() len = %v, want %v", len(got), tt.wantLen)
			}

			// Verify paths are constructed correctly
			for _, folder := range got {
				expectedReqPath := tt.requestDir + "/" + folder.KassaCode + "/" + folder.FolderName
				if folder.RequestPath != expectedReqPath {
					t.Errorf("GetAllKassaFolders() RequestPath = %v, want %v", folder.RequestPath, expectedReqPath)
				}

				expectedRespPath := tt.responseDir + "/" + folder.KassaCode + "/" + folder.FolderName
				if folder.ResponsePath != expectedRespPath {
					t.Errorf("GetAllKassaFolders() ResponsePath = %v, want %v", folder.ResponsePath, expectedRespPath)
				}
			}
		})
	}
}

func TestRetryOperation(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		failCount  int // How many times the operation should fail before succeeding
		wantErr    bool
	}{
		{
			name:       "succeeds immediately",
			maxRetries: 3,
			failCount:  0,
			wantErr:    false,
		},
		{
			name:       "succeeds after one retry",
			maxRetries: 3,
			failCount:  1,
			wantErr:    false,
		},
		{
			name:       "succeeds on last try",
			maxRetries: 3,
			failCount:  2,
			wantErr:    false,
		},
		{
			name:       "fails after all retries",
			maxRetries: 3,
			failCount:  5,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			operation := func() error {
				callCount++
				if callCount <= tt.failCount {
					return os.ErrNotExist // Simulate error
				}
				return nil
			}

			err := RetryOperation(operation, tt.maxRetries, 1*time.Millisecond)

			if tt.wantErr {
				if err == nil {
					t.Error("RetryOperation() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("RetryOperation() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRetryOperationContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	callCount := 0
	operation := func() error {
		callCount++
		if callCount == 1 {
			cancel()
			return os.ErrNotExist
		}
		return ctx.Err()
	}

	err := RetryOperation(operation, 2, 1*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("RetryOperation() expected context canceled error, got %v", err)
	}
}

func TestRetryOperationNonPositiveDelay(t *testing.T) {
	tests := []struct {
		name  string
		delay time.Duration
	}{
		{name: "zero_delay", delay: 0},
		{name: "negative_delay", delay: -1 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			operation := func() error {
				callCount++
				if callCount == 1 {
					return os.ErrNotExist
				}
				return nil
			}

			if err := RetryOperation(operation, 2, tt.delay); err != nil {
				t.Fatalf("RetryOperation() unexpected error: %v", err)
			}
		})
	}
}

// Benchmark for CreateRequestFile
func BenchmarkCreateRequestFile(b *testing.B) {
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subDir := filepath.Join(tmpDir, "bench", string(rune(i%1000)))
		_, _ = CreateRequestFile(subDir, "2024-12-01")
	}
}
