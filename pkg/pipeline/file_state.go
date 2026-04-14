package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/user/go-frontol-loader/pkg/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

type fileLifecycleStage string

const (
	fileLifecycleStagePendingFinalize fileLifecycleStage = "pending_finalize"
	fileLifecycleStageParseFailed     fileLifecycleStage = "parse_failed"
	fileLifecycleStageCompleted       fileLifecycleStage = "completed"
)

type fileLifecycleRecord struct {
	Key                 string             `json:"key"`
	LogicalKey          string             `json:"logical_key"`
	RemotePath          string             `json:"remote_path"`
	RequestedDate       string             `json:"requested_date,omitempty"`
	Filename            string             `json:"filename"`
	SourceFolder        string             `json:"source_folder"`
	DBID                string             `json:"db_id,omitempty"`
	ReportNum           string             `json:"report_num,omitempty"`
	ContentHash         string             `json:"content_hash"`
	TransactionCount    int                `json:"transaction_count"`
	TransactionManifest map[string][]int64 `json:"transaction_manifest,omitempty"`
	Stage               fileLifecycleStage `json:"stage"`
	UpdatedAt           time.Time          `json:"updated_at"`
	LastError           string             `json:"last_error,omitempty"`
}

type fileLifecycleStore struct {
	root string
}

func newFileLifecycleStore(baseDir string) *fileLifecycleStore {
	return &fileLifecycleStore{root: filepath.Join(baseDir, ".etl-state")}
}

func newFileLifecycleRecord(store *fileLifecycleStore, logicalKey string, remotePath string, requestedDate string, filename string, sourceFolder string, header *models.FileHeader, contentHash string, transactionCount int) *fileLifecycleRecord {
	record := &fileLifecycleRecord{
		Key:              store.key(logicalKey, contentHash),
		LogicalKey:       logicalKey,
		RemotePath:       remotePath,
		RequestedDate:    requestedDate,
		Filename:         filename,
		SourceFolder:     sourceFolder,
		ContentHash:      contentHash,
		TransactionCount: transactionCount,
		Stage:            fileLifecycleStagePendingFinalize,
		UpdatedAt:        time.Now(),
	}
	if header != nil {
		record.DBID = header.DBID
		record.ReportNum = header.ReportNum
	}
	return record
}

func (r *fileLifecycleRecord) withManifest(manifest map[string][]int64) *fileLifecycleRecord {
	clone := *r
	clone.TransactionManifest = cloneManifest(manifest)
	clone.UpdatedAt = time.Now()
	return &clone
}

func (r *fileLifecycleRecord) withStage(stage fileLifecycleStage, lastError string) *fileLifecycleRecord {
	clone := *r
	clone.Stage = stage
	clone.LastError = lastError
	clone.UpdatedAt = time.Now()
	return &clone
}

func (s *fileLifecycleStore) key(logicalKey string, contentHash string) string {
	sum := sha256.Sum256([]byte(logicalKey + "|" + contentHash))
	return hex.EncodeToString(sum[:])
}

func (s *fileLifecycleStore) pathFor(key string) string {
	return filepath.Join(s.root, key+".json")
}

func (s *fileLifecycleStore) latestPathFor(logicalKey string) string {
	sum := sha256.Sum256([]byte(logicalKey))
	return filepath.Join(s.root, "latest", hex.EncodeToString(sum[:])+".json")
}

func (s *fileLifecycleStore) Load(key string) (*fileLifecycleRecord, error) {
	path := s.pathFor(key)
	// #nosec G304 -- path is derived from controlled LOCAL_DIR/.etl-state paths.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read lifecycle state: %w", err)
	}

	var record fileLifecycleRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("decode lifecycle state: %w", err)
	}
	return &record, nil
}

func (s *fileLifecycleStore) Save(record *fileLifecycleRecord) error {
	if err := os.MkdirAll(s.root, 0750); err != nil {
		return fmt.Errorf("create lifecycle directory: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("encode lifecycle state: %w", err)
	}

	path := s.pathFor(record.Key)
	tempPath := path + ".tmp"
	// #nosec G306 -- state files are internal process artifacts under LOCAL_DIR/.etl-state.
	if err := os.WriteFile(tempPath, data, 0640); err != nil {
		return fmt.Errorf("write lifecycle state: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("persist lifecycle state: %w", err)
	}
	return nil
}

func (s *fileLifecycleStore) SaveLatest(record *fileLifecycleRecord) error {
	latestDir := filepath.Dir(s.latestPathFor(record.LogicalKey))
	if err := os.MkdirAll(latestDir, 0750); err != nil {
		return fmt.Errorf("create lifecycle latest directory: %w", err)
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("encode latest lifecycle state: %w", err)
	}
	path := s.latestPathFor(record.LogicalKey)
	tempPath := path + ".tmp"
	// #nosec G306 -- state files are internal process artifacts under LOCAL_DIR/.etl-state.
	if err := os.WriteFile(tempPath, data, 0640); err != nil {
		return fmt.Errorf("write latest lifecycle state: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("persist latest lifecycle state: %w", err)
	}
	return nil
}

func (s *fileLifecycleStore) LoadLatest(logicalKey string) (*fileLifecycleRecord, error) {
	path := s.latestPathFor(logicalKey)
	// #nosec G304 -- path is derived from controlled LOCAL_DIR/.etl-state paths.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read latest lifecycle state: %w", err)
	}
	var record fileLifecycleRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("decode latest lifecycle state: %w", err)
	}
	return &record, nil
}

func hashLocalFile(path string) (string, error) {
	// #nosec G304 -- path is derived from controlled LOCAL_DIR/.etl-state paths.
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file for hash: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func finalizeFileLifecycle(ctx context.Context, ftpClient ftp.FTPClient, store *fileLifecycleStore, record *fileLifecycleRecord, logger *slog.Logger) error {
	if err := ftpClient.MarkFileAsProcessed(record.RemotePath); err != nil {
		logger.ErrorContext(ctx, "Failed to mark file as processed",
			"file", record.RemotePath,
			"error", err.Error(),
			"event", "mark_processed_error",
		)
		if saveErr := store.Save(record.withStage(fileLifecycleStagePendingFinalize, err.Error())); saveErr != nil {
			return fmt.Errorf("failed to mark file as processed: %w (also failed to persist lifecycle state: %v)", err, saveErr)
		}
		return fmt.Errorf("failed to mark file as processed: %w", err)
	}

	logger.DebugContext(ctx, "File marked as processed",
		"file", record.RemotePath,
		"event", "file_marked_processed",
	)

	if err := store.Save(record.withStage(fileLifecycleStageCompleted, "")); err != nil {
		logger.WarnContext(ctx, "Failed to persist completed lifecycle state",
			"file", record.RemotePath,
			"error", err.Error(),
			"event", "file_lifecycle_persist_warning",
		)
	}
	if err := store.SaveLatest(record.withStage(fileLifecycleStageCompleted, "")); err != nil {
		logger.WarnContext(ctx, "Failed to persist latest lifecycle state",
			"file", record.RemotePath,
			"error", err.Error(),
			"event", "file_lifecycle_latest_persist_warning",
		)
	}

	return nil
}

func cloneManifest(manifest map[string][]int64) map[string][]int64 {
	if len(manifest) == 0 {
		return nil
	}
	cloned := make(map[string][]int64, len(manifest))
	for table, ids := range manifest {
		copied := make([]int64, len(ids))
		copy(copied, ids)
		cloned[table] = copied
	}
	return cloned
}
