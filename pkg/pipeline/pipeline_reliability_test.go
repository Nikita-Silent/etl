package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	ftplib "github.com/jlaffaye/ftp"
	ftpclient "github.com/user/go-frontol-loader/pkg/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

type mockFileLoader struct {
	getTransactionCount       func(map[string]interface{}) int
	loadFileData              func(context.Context, map[string]interface{}) error
	loadFileDataWithReconcile func(context.Context, *models.FileLoadState, map[string][]int64, map[string]interface{}) error
	getFileLoadState          func(context.Context, string) (*models.FileLoadState, error)
	getDetails                func(map[string]interface{}) []map[string]interface{}
}

func (m *mockFileLoader) GetTransactionCount(transactions map[string]interface{}) int {
	if m.getTransactionCount != nil {
		return m.getTransactionCount(transactions)
	}
	return 0
}

func (m *mockFileLoader) LoadFileData(ctx context.Context, transactions map[string]interface{}) error {
	if m.loadFileData != nil {
		return m.loadFileData(ctx, transactions)
	}
	return nil
}

func (m *mockFileLoader) LoadFileDataWithReconcile(ctx context.Context, fileState *models.FileLoadState, staleManifest map[string][]int64, transactions map[string]interface{}) error {
	if m.loadFileDataWithReconcile != nil {
		return m.loadFileDataWithReconcile(ctx, fileState, staleManifest, transactions)
	}
	return m.LoadFileData(ctx, transactions)
}

func (m *mockFileLoader) GetFileLoadState(ctx context.Context, logicalKey string) (*models.FileLoadState, error) {
	if m.getFileLoadState != nil {
		return m.getFileLoadState(ctx, logicalKey)
	}
	return nil, nil
}

func (m *mockFileLoader) GetTransactionDetails(transactions map[string]interface{}) []map[string]interface{} {
	if m.getDetails != nil {
		return m.getDetails(transactions)
	}
	return nil
}

func TestRunWithClientsMarksPartialStatusOnOperationalIssues(t *testing.T) {
	oldProcess := processFilesFromFTPFunc
	defer func() {
		processFilesFromFTPFunc = oldProcess
	}()

	processFilesFromFTPFunc = func(ctx context.Context, ftpClient ftpclient.FTPClient, loader fileLoader, cfg *models.Config, date string, logger *slog.Logger) (*ProcessingStats, error) {
		return &ProcessingStats{
			FilesProcessed:     1,
			TransactionsLoaded: 12,
			KassaDetails: []KassaProcessingStats{{
				KassaCode:      "P13",
				FolderName:     "P13",
				SourceFolder:   "P13/P13",
				FilesFound:     2,
				FilesQueued:    1,
				FilesProcessed: 1,
			}},
			ErrorBreakdown: map[string]int{
				"file_process_error": 1,
			},
			ErrorSamples: []PipelineIssueSample{{Stage: "file_process_error", File: "broken.txt", Error: "boom"}},
		}, nil
	}

	ftpMock := &ftpclient.MockClient{}

	result, err := runWithClients(
		context.Background(),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		&models.Config{WaitDelayMinutes: time.Millisecond},
		"2026-04-13",
		ftpMock,
		&mockFileLoader{},
		&PipelineResult{StartTime: time.Now(), Date: "2026-04-13", Status: PipelineStatusFailed},
	)
	if err != nil {
		t.Fatalf("runWithClients() unexpected error: %v", err)
	}
	if result.Status != PipelineStatusPartial {
		t.Fatalf("runWithClients() status = %q, want %q", result.Status, PipelineStatusPartial)
	}
	if result.Success {
		t.Fatal("runWithClients() success = true, want false for partial result")
	}
	if result.Errors != 1 {
		t.Fatalf("runWithClients() errors = %d, want %d", result.Errors, 1)
	}
	if result.ErrorBreakdown["file_process_error"] != 1 {
		t.Fatalf("runWithClients() missing file_process_error breakdown: %+v", result.ErrorBreakdown)
	}
	if result.ErrorMessage == "" {
		t.Fatal("runWithClients() error message should summarize partial result")
	}
	if len(result.KassaDetails) != 1 || result.KassaDetails[0].SourceFolder != "P13/P13" {
		t.Fatalf("runWithClients() kassa details = %+v, want propagated stats", result.KassaDetails)
	}
}

func TestProcessFilePersistsPendingFinalizationState(t *testing.T) {
	localDir := t.TempDir()
	folder := models.KassaFolder{
		KassaCode:    "P13",
		FolderName:   "P13",
		ResponsePath: "/response/P13",
	}
	samplePath := filepath.Join(findRepoRoot(t), "data", "response.txt")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	loadCalls := 0

	ftpMock := &ftpclient.MockClient{
		DownloadFileFunc: func(remotePath, localPath string) error {
			return copyFile(samplePath, localPath)
		},
		MarkFileAsProcessedFunc: func(remotePath string) error {
			return errors.New("rename failed")
		},
	}

	loader := &mockFileLoader{
		getTransactionCount: func(transactions map[string]interface{}) int { return 1346 },
		loadFileData: func(ctx context.Context, transactions map[string]interface{}) error {
			loadCalls++
			return nil
		},
		getDetails: func(transactions map[string]interface{}) []map[string]interface{} {
			return []map[string]interface{}{{"table_name": "tx_document_open_42", "count": 82}}
		},
	}

	_, err := processFile(context.Background(), ftpMock, loader, &models.Config{LocalDir: localDir}, "response.txt", folder, "2024-12-01", logger)
	if err == nil {
		t.Fatal("processFile() expected error when mark processed fails")
	}
	if loadCalls != 1 {
		t.Fatalf("processFile() load calls = %d, want %d", loadCalls, 1)
	}

	hash, err := hashLocalFile(samplePath)
	if err != nil {
		t.Fatalf("hashLocalFile() unexpected error: %v", err)
	}
	store := newFileLifecycleStore(localDir)
	record, err := store.Load(store.key("/response/P13/response.txt|2024-12-01", hash))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if record == nil {
		t.Fatal("Load() returned nil record")
	}
	if record.Stage != fileLifecycleStagePendingFinalize {
		t.Fatalf("record stage = %q, want %q", record.Stage, fileLifecycleStagePendingFinalize)
	}
	if record.TransactionCount != 1346 {
		t.Fatalf("record transaction count = %d, want %d", record.TransactionCount, 1346)
	}
}

func TestProcessFileRecoversPendingFinalizationWithoutReload(t *testing.T) {
	localDir := t.TempDir()
	folder := models.KassaFolder{
		KassaCode:    "P13",
		FolderName:   "P13",
		ResponsePath: "/response/P13",
	}
	samplePath := filepath.Join(findRepoRoot(t), "data", "response.txt")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	loadCalls := 0
	markCalls := 0

	hash, err := hashLocalFile(samplePath)
	if err != nil {
		t.Fatalf("hashLocalFile() unexpected error: %v", err)
	}
	store := newFileLifecycleStore(localDir)
	record := &fileLifecycleRecord{
		Key:              store.key("/response/P13/response.txt|2024-12-01", hash),
		LogicalKey:       "/response/P13/response.txt|2024-12-01",
		RemotePath:       "/response/P13/response.txt",
		RequestedDate:    "2024-12-01",
		Filename:         "response.txt",
		SourceFolder:     "P13/P13",
		DBID:             "1",
		ReportNum:        "24335",
		ContentHash:      hash,
		TransactionCount: 1346,
		Stage:            fileLifecycleStagePendingFinalize,
		UpdatedAt:        time.Now(),
	}
	if err := store.Save(record); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	ftpMock := &ftpclient.MockClient{
		DownloadFileFunc: func(remotePath, localPath string) error {
			return copyFile(samplePath, localPath)
		},
		MarkFileAsProcessedFunc: func(remotePath string) error {
			markCalls++
			return nil
		},
	}

	loader := &mockFileLoader{
		getTransactionCount: func(transactions map[string]interface{}) int { return 1346 },
		loadFileData: func(ctx context.Context, transactions map[string]interface{}) error {
			loadCalls++
			return nil
		},
	}

	outcome, err := processFile(context.Background(), ftpMock, loader, &models.Config{LocalDir: localDir}, "response.txt", folder, "2024-12-01", logger)
	if err != nil {
		t.Fatalf("processFile() unexpected error: %v", err)
	}
	if loadCalls != 0 {
		t.Fatalf("processFile() load calls = %d, want %d", loadCalls, 0)
	}
	if markCalls != 1 {
		t.Fatalf("processFile() mark calls = %d, want %d", markCalls, 1)
	}
	if !outcome.Recovered {
		t.Fatal("processFile() should report recovered lifecycle state")
	}
	if outcome.LoadedTransactions != 0 {
		t.Fatalf("processFile() loaded transactions = %d, want 0 during recovery", outcome.LoadedTransactions)
	}

	updated, err := store.Load(record.Key)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if updated == nil {
		t.Fatal("Load() returned nil updated record")
	}
	if updated.Stage != fileLifecycleStageCompleted {
		t.Fatalf("updated record stage = %q, want %q", updated.Stage, fileLifecycleStageCompleted)
	}
}

func TestProcessFileQuarantinesRepeatedParseFailures(t *testing.T) {
	localDir := t.TempDir()
	folder := models.KassaFolder{
		KassaCode:    "P13",
		FolderName:   "P13",
		ResponsePath: "/response/P13",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	badContent := []byte("1\nDB_TEST\nREPORT_001\nnot-a-number;01.12.2024;10:30:00;1;1;100;1\n")
	downloadCalls := 0
	loadCalls := 0

	ftpMock := &ftpclient.MockClient{
		DownloadFileFunc: func(remotePath, localPath string) error {
			downloadCalls++
			if err := os.MkdirAll(filepath.Dir(localPath), 0750); err != nil {
				return err
			}
			return os.WriteFile(localPath, badContent, 0640)
		},
	}

	loader := &mockFileLoader{
		loadFileData: func(ctx context.Context, transactions map[string]interface{}) error {
			loadCalls++
			return nil
		},
	}

	_, err := processFile(context.Background(), ftpMock, loader, &models.Config{LocalDir: localDir}, "broken.txt", folder, "2024-12-01", logger)
	if err == nil {
		t.Fatal("processFile() expected parse error, got nil")
	}
	if stage := stageForFileError(err); stage != "file_parse_error" {
		t.Fatalf("stageForFileError() = %q, want %q", stage, "file_parse_error")
	}
	if loadCalls != 0 {
		t.Fatalf("processFile() load calls = %d, want 0", loadCalls)
	}

	hash := sha256.Sum256(badContent)
	store := newFileLifecycleStore(localDir)
	record, err := store.Load(store.key("/response/P13/broken.txt|2024-12-01", hex.EncodeToString(hash[:])))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if record == nil || record.Stage != fileLifecycleStageParseFailed {
		t.Fatalf("expected parse_failed lifecycle record, got %#v", record)
	}

	_, err = processFile(context.Background(), ftpMock, loader, &models.Config{LocalDir: localDir}, "broken.txt", folder, "2024-12-01", logger)
	if err == nil {
		t.Fatal("processFile() expected quarantined error on repeated parse failure")
	}
	if stage := stageForFileError(err); stage != "file_quarantined" {
		t.Fatalf("stageForFileError() = %q, want %q", stage, "file_quarantined")
	}
	if downloadCalls != 2 {
		t.Fatalf("download calls = %d, want %d", downloadCalls, 2)
	}
	if loadCalls != 0 {
		t.Fatalf("processFile() load calls after quarantine = %d, want 0", loadCalls)
	}
}

func TestProcessFileReconcilesCorrectedReupload(t *testing.T) {
	localDir := t.TempDir()
	folder := models.KassaFolder{
		KassaCode:    "P13",
		FolderName:   "P13",
		ResponsePath: "/response/P13",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	originalPath := filepath.Join(findRepoRoot(t), "data", "response.txt")
	originalBytes, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}
	correctedBytes := append([]byte{}, originalBytes...)
	correctedBytes = append(correctedBytes, []byte("\n")...)
	oldHash := sha256.Sum256(originalBytes)
	reconcileCalls := 0
	var gotManifest map[string][]int64
	var gotSourceFolder string

	ftpMock := &ftpclient.MockClient{
		DownloadFileFunc: func(remotePath, localPath string) error {
			if err := os.MkdirAll(filepath.Dir(localPath), 0750); err != nil {
				return err
			}
			return os.WriteFile(localPath, correctedBytes, 0640)
		},
		MarkFileAsProcessedFunc: func(remotePath string) error { return nil },
	}

	loader := &mockFileLoader{
		getTransactionCount: func(transactions map[string]interface{}) int { return 1346 },
		getFileLoadState: func(ctx context.Context, logicalKey string) (*models.FileLoadState, error) {
			return &models.FileLoadState{
				LogicalKey:    logicalKey,
				RemotePath:    "/response/P13/response.txt",
				RequestedDate: "2024-12-01",
				SourceFolder:  "P13/P13",
				ContentHash:   hex.EncodeToString(oldHash[:]),
				TransactionManifest: map[string][]int64{
					"tx_document_open_42": {733685},
				},
			}, nil
		},
		loadFileDataWithReconcile: func(ctx context.Context, fileState *models.FileLoadState, staleManifest map[string][]int64, transactions map[string]interface{}) error {
			reconcileCalls++
			gotSourceFolder = fileState.SourceFolder
			gotManifest = staleManifest
			return nil
		},
		getDetails: func(transactions map[string]interface{}) []map[string]interface{} {
			return []map[string]interface{}{{"table_name": "tx_document_open_42", "count": 82}}
		},
	}

	_, err = processFile(context.Background(), ftpMock, loader, &models.Config{LocalDir: localDir}, "response.txt", folder, "2024-12-01", logger)
	if err != nil {
		t.Fatalf("processFile() unexpected error: %v", err)
	}
	if reconcileCalls != 1 {
		t.Fatalf("LoadFileDataWithReconcile() calls = %d, want 1", reconcileCalls)
	}
	if gotSourceFolder != "P13/P13" {
		t.Fatalf("source folder = %q, want %q", gotSourceFolder, "P13/P13")
	}
	if len(gotManifest["tx_document_open_42"]) != 1 || gotManifest["tx_document_open_42"][0] != 733685 {
		t.Fatalf("stale manifest = %+v, want previous record manifest", gotManifest)
	}
}

func TestProcessFileDoesNotReconcileDifferentRequestedDates(t *testing.T) {
	localDir := t.TempDir()
	folder := models.KassaFolder{
		KassaCode:    "P13",
		FolderName:   "P13",
		ResponsePath: "/response/P13",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	originalPath := filepath.Join(findRepoRoot(t), "data", "response.txt")
	originalBytes, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}
	reconcileCalls := 0
	ftpMock := &ftpclient.MockClient{
		DownloadFileFunc: func(remotePath, localPath string) error {
			if err := os.MkdirAll(filepath.Dir(localPath), 0750); err != nil {
				return err
			}
			return os.WriteFile(localPath, originalBytes, 0640)
		},
		MarkFileAsProcessedFunc: func(remotePath string) error { return nil },
	}
	loader := &mockFileLoader{
		getTransactionCount: func(transactions map[string]interface{}) int { return 1346 },
		getFileLoadState: func(ctx context.Context, logicalKey string) (*models.FileLoadState, error) {
			return nil, nil
		},
		loadFileDataWithReconcile: func(ctx context.Context, fileState *models.FileLoadState, staleManifest map[string][]int64, transactions map[string]interface{}) error {
			reconcileCalls++
			if len(staleManifest) != 0 {
				t.Fatalf("staleManifest = %+v, want empty for different requested date", staleManifest)
			}
			return nil
		},
	}

	_, err = processFile(context.Background(), ftpMock, loader, &models.Config{LocalDir: localDir}, "response.txt", folder, "2024-12-02", logger)
	if err != nil {
		t.Fatalf("processFile() unexpected error: %v", err)
	}
	if reconcileCalls != 1 {
		t.Fatalf("LoadFileDataWithReconcile() calls = %d, want 1", reconcileCalls)
	}
}

func TestProcessFileUsesDurableStateWhenLocalStateSaveFails(t *testing.T) {
	localDir := t.TempDir()
	folder := models.KassaFolder{
		KassaCode:    "P13",
		FolderName:   "P13",
		ResponsePath: "/response/P13",
	}
	samplePath := filepath.Join(findRepoRoot(t), "data", "response.txt")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	loadCalls := 0
	markCalls := 0
	blockingStatePath := filepath.Join(localDir, ".etl-state")
	if err := os.MkdirAll(blockingStatePath, 0500); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(blockingStatePath, 0750) })

	ftpMock := &ftpclient.MockClient{
		DownloadFileFunc: func(remotePath, localPath string) error {
			return copyFile(samplePath, localPath)
		},
		MarkFileAsProcessedFunc: func(remotePath string) error {
			markCalls++
			return nil
		},
	}

	loader := &mockFileLoader{
		getTransactionCount: func(transactions map[string]interface{}) int { return 1346 },
		loadFileDataWithReconcile: func(ctx context.Context, fileState *models.FileLoadState, staleManifest map[string][]int64, transactions map[string]interface{}) error {
			loadCalls++
			return nil
		},
		getFileLoadState: func(ctx context.Context, logicalKey string) (*models.FileLoadState, error) {
			if loadCalls == 0 {
				return nil, nil
			}
			hash, err := hashLocalFile(samplePath)
			if err != nil {
				t.Fatalf("hashLocalFile() unexpected error: %v", err)
			}
			return &models.FileLoadState{
				LogicalKey:    logicalKey,
				RemotePath:    "/response/P13/response.txt",
				RequestedDate: "2024-12-01",
				SourceFolder:  "P13/P13",
				ContentHash:   hash,
			}, nil
		},
	}

	cfg := &models.Config{LocalDir: localDir}
	_, err := processFile(context.Background(), ftpMock, loader, cfg, "response.txt", folder, "2024-12-01", logger)
	if err == nil || stageForFileError(err) != "file_state_save_error" {
		t.Fatalf("processFile() error = %v, want file_state_save_error", err)
	}
	if loadCalls != 1 {
		t.Fatalf("first processFile() load calls = %d, want 1", loadCalls)
	}

	_, err = processFile(context.Background(), ftpMock, loader, &models.Config{LocalDir: localDir}, "response.txt", folder, "2024-12-01", logger)
	if err != nil {
		t.Fatalf("second processFile() unexpected error: %v", err)
	}
	if loadCalls != 1 {
		t.Fatalf("second processFile() load calls = %d, want still 1", loadCalls)
	}
	if markCalls != 1 {
		t.Fatalf("mark processed calls = %d, want 1", markCalls)
	}
}

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0640)
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
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

var _ fileLoader = (*mockFileLoader)(nil)
var _ ftpclient.FTPClient = (*ftpclient.MockClient)(nil)
var _ = ftplib.Entry{}
