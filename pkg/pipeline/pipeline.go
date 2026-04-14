package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sort"
	"sync"
	"time"

	ftplib "github.com/jlaffaye/ftp"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
	"github.com/user/go-frontol-loader/pkg/parser"
	"github.com/user/go-frontol-loader/pkg/repository"
)

type fileLoader interface {
	GetTransactionCount(transactions map[string]interface{}) int
	LoadFileData(ctx context.Context, transactions map[string]interface{}) error
	LoadFileDataWithReconcile(ctx context.Context, sourceFolder string, staleManifest map[string][]int64, transactions map[string]interface{}) error
	GetTransactionDetails(transactions map[string]interface{}) []map[string]interface{}
}

type PipelineStatus string

const (
	PipelineStatusCompleted PipelineStatus = "completed"
	PipelineStatusPartial   PipelineStatus = "partial"
	PipelineStatusFailed    PipelineStatus = "failed"
)

const maxErrorSamples = 5

var processFilesFromFTPFunc func(context.Context, ftp.FTPClient, fileLoader, *models.Config, string, *slog.Logger) (*ProcessingStats, error) = processFilesFromFTP

type PipelineIssueSample struct {
	Stage string `json:"stage"`
	File  string `json:"file,omitempty"`
	Path  string `json:"path,omitempty"`
	Error string `json:"error,omitempty"`
}

// TransactionTypeStats представляет статистику по типу транзакций
type TransactionTypeStats struct {
	TableName string `json:"table_name"`
	Count     int    `json:"count"`
}

type KassaProcessingStats struct {
	KassaCode        string `json:"kassa_code"`
	FolderName       string `json:"folder_name"`
	SourceFolder     string `json:"source_folder"`
	Status           string `json:"status,omitempty"`
	RequestPath      string `json:"request_path,omitempty"`
	ResponsePath     string `json:"response_path"`
	FilesFound       int    `json:"files_found"`
	FilesQueued      int    `json:"files_queued"`
	FilesProcessed   int    `json:"files_processed"`
	FilesSkipped     int    `json:"files_skipped"`
	FilesRecovered   int    `json:"files_recovered,omitempty"`
	FilesFailed      int    `json:"files_failed,omitempty"`
	DeletedRequests  int    `json:"deleted_requests,omitempty"`
	DeletedResponses int    `json:"deleted_responses,omitempty"`
	LockWait         string `json:"lock_wait,omitempty"`
	LastIssueStage   string `json:"last_issue_stage,omitempty"`
	LastIssueMessage string `json:"last_issue_message,omitempty"`
}

// PipelineResult содержит результат выполнения ETL-конвейера
type PipelineResult struct {
	StartTime          time.Time              `json:"start_time"`
	EndTime            time.Time              `json:"end_time"`
	Duration           string                 `json:"duration"`
	Date               string                 `json:"date"`
	Status             PipelineStatus         `json:"status"`
	FilesProcessed     int                    `json:"files_processed"`
	FilesSkipped       int                    `json:"files_skipped"`
	FilesRecovered     int                    `json:"files_recovered,omitempty"`
	TransactionsLoaded int                    `json:"transactions_loaded"`
	Errors             int                    `json:"errors"`
	Success            bool                   `json:"success"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	ErrorBreakdown     map[string]int         `json:"error_breakdown,omitempty"`
	ErrorSamples       []PipelineIssueSample  `json:"error_samples,omitempty"`
	KassaDetails       []KassaProcessingStats `json:"kassa_details,omitempty"`
	TransactionDetails []TransactionTypeStats `json:"transaction_details,omitempty"` // Детальная информация по типам транзакций
}

// Run выполняет полный ETL-конвейер для указанной даты
func Run(ctx context.Context, logger *slog.Logger, cfg *models.Config, date string) (*PipelineResult, error) {
	result := &PipelineResult{
		StartTime: time.Now(),
		Date:      date,
		Status:    PipelineStatusFailed,
		Success:   false,
	}
	defer finalizeResult(result)

	logger.InfoContext(ctx, "Starting ETL pipeline",
		"log_kind", "loki_operational",
		"date", date,
		"kassa_count", len(cfg.KassaStructure),
		"ftp_pool_size", cfg.FTPPoolSize,
		"worker_pool_size", cfg.WorkerPoolSize,
		"event", "etl_start",
	)

	// Инициализация базы данных
	database, err := db.NewPool(cfg)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to connect to database: %v", err)
		logger.ErrorContext(ctx, "Failed to connect to database",
			"error", err.Error(),
			"event", "db_connection_failed",
		)
		return result, err
	}
	defer database.Close()

	// Инициализация FTP connection pool
	ftpClient, err := ftp.NewPool(cfg, cfg.FTPPoolSize)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create FTP pool: %v", err)
		logger.ErrorContext(ctx, "Failed to create FTP pool",
			"error", err.Error(),
			"event", "ftp_pool_creation_failed",
		)
		return result, err
	}
	defer func() {
		if err := ftpClient.Close(); err != nil {
			logger.WarnContext(ctx, "Failed to close FTP pool",
				"error", err.Error(),
				"event", "ftp_pool_close_error",
			)
		}
	}()

	// Инициализация загрузчика
	loader := repository.NewLoader(database)

	return runWithClients(ctx, logger, cfg, date, ftpClient, loader, result)

}

func runWithClients(ctx context.Context, logger *slog.Logger, cfg *models.Config, date string, ftpClient ftp.FTPClient, loader fileLoader, result *PipelineResult) (*PipelineResult, error) {
	issues := newIssueCollector()

	// Шаг 1: Координация загрузки по каждой папке отдельно.
	logger.InfoContext(ctx, "Step 1: Coordinating per-folder FTP lifecycle",
		"event", "etl_step_4",
	)
	logger.DebugContext(ctx, "FTP configuration",
		"ftp_request_dir", cfg.FTPRequestDir,
		"ftp_response_dir", cfg.FTPResponseDir,
		"kassa_structure", fmt.Sprintf("%v", cfg.KassaStructure),
	)

	stats, err := processFilesFromFTPFunc(ctx, ftpClient, loader, cfg, date, logger)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to process files from FTP: %v", err)
		issues.Merge(stats)
		result.Errors = issues.Total()
		result.ErrorBreakdown = issues.CloneBreakdown()
		result.ErrorSamples = issues.CloneSamples()
		logger.ErrorContext(ctx, "Failed to process files from FTP",
			"error", err.Error(),
			"event", "etl_step_4_error",
		)
		return result, err
	}

	// Обновляем результат
	result.FilesProcessed = stats.FilesProcessed
	result.FilesSkipped = stats.FilesSkipped
	result.FilesRecovered = stats.FilesRecovered
	result.TransactionsLoaded = stats.TransactionsLoaded
	issues.Merge(stats)
	result.Errors = issues.Total()
	result.ErrorBreakdown = issues.CloneBreakdown()
	result.ErrorSamples = issues.CloneSamples()
	result.KassaDetails = stats.KassaDetails
	result.TransactionDetails = stats.TransactionDetails
	result.Status = PipelineStatusCompleted
	if result.Errors > 0 {
		result.Status = PipelineStatusPartial
		result.ErrorMessage = issues.Summary()
	}
	result.Success = result.Status == PipelineStatusCompleted

	logMethod := logger.InfoContext
	message := "ETL pipeline completed successfully"
	if result.Status == PipelineStatusPartial {
		logMethod = logger.WarnContext
		message = "ETL pipeline completed with recoverable errors"
	}
	logMethod(ctx, message,
		"log_kind", "loki_operational",
		"date", date,
		"status", result.Status,
		"files_processed", result.FilesProcessed,
		"files_skipped", result.FilesSkipped,
		"files_recovered", result.FilesRecovered,
		"transactions_loaded", result.TransactionsLoaded,
		"errors", result.Errors,
		"kassa_details", result.KassaDetails,
		"event", "etl_complete",
	)
	logger.InfoContext(ctx, "ETL run summary",
		"log_kind", "loki_operational",
		"date", date,
		"status", result.Status,
		"duration_ms", result.EndTime.Sub(result.StartTime).Milliseconds(),
		"files_processed", result.FilesProcessed,
		"files_skipped", result.FilesSkipped,
		"files_recovered", result.FilesRecovered,
		"transactions_loaded", result.TransactionsLoaded,
		"errors", result.Errors,
		"kassa_count", len(result.KassaDetails),
		"event", "etl_run_summary",
	)
	for _, detail := range result.KassaDetails {
		if detail.FilesFound == 0 && detail.FilesProcessed == 0 && detail.FilesFailed == 0 && detail.FilesRecovered == 0 {
			continue
		}
		logger.InfoContext(ctx, "ETL kassa summary",
			"log_kind", "loki_operational",
			"date", date,
			"kassa_code", detail.KassaCode,
			"folder_name", detail.FolderName,
			"source_folder", detail.SourceFolder,
			"status", detail.Status,
			"files_found", detail.FilesFound,
			"files_queued", detail.FilesQueued,
			"files_processed", detail.FilesProcessed,
			"files_skipped", detail.FilesSkipped,
			"files_recovered", detail.FilesRecovered,
			"files_failed", detail.FilesFailed,
			"deleted_requests", detail.DeletedRequests,
			"deleted_responses", detail.DeletedResponses,
			"lock_wait", detail.LockWait,
			"last_issue_stage", detail.LastIssueStage,
			"event", "etl_kassa_summary",
		)
	}

	return result, nil
}

// ProcessingStats содержит статистику обработки файлов
type ProcessingStats struct {
	FilesProcessed     int
	FilesSkipped       int
	FilesRecovered     int
	TransactionsLoaded int
	Errors             int
	ErrorBreakdown     map[string]int
	ErrorSamples       []PipelineIssueSample
	KassaDetails       []KassaProcessingStats
	TransactionDetails []TransactionTypeStats // Детальная статистика по типам транзакций
}

type folderRunResult struct {
	Detail             KassaProcessingStats
	ErrorBreakdown     map[string]int
	ErrorSamples       []PipelineIssueSample
	TransactionDetails map[string]int
}

// fileTask сохранен для обратной совместимости тестов пакета pipeline.
type fileTask struct {
	folder models.KassaFolder
	file   *ftplib.Entry
}

type fileProcessOutcome struct {
	LoadedTransactions int
	TransactionDetails []map[string]interface{}
	Recovered          bool
}

type stagedFileError struct {
	stage string
	err   error
}

func (e *stagedFileError) Error() string {
	return e.err.Error()
}

func (e *stagedFileError) Unwrap() error {
	return e.err
}

func newStagedFileError(stage string, err error) error {
	if err == nil {
		return nil
	}
	return &stagedFileError{stage: stage, err: err}
}

func stageForFileError(err error) string {
	var staged *stagedFileError
	if errors.As(err, &staged) {
		return staged.stage
	}
	return "file_process_error"
}

// processFilesFromFTP координирует полный цикл загрузки по каждой папке.
func processFilesFromFTP(ctx context.Context, ftpClient ftp.FTPClient, loader fileLoader, cfg *models.Config, date string, logger *slog.Logger) (*ProcessingStats, error) {
	folders := ftp.GetAllKassaFolders(cfg)

	stats := &ProcessingStats{
		TransactionDetails: make([]TransactionTypeStats, 0),
		ErrorBreakdown:     make(map[string]int),
		ErrorSamples:       make([]PipelineIssueSample, 0, maxErrorSamples),
		KassaDetails:       make([]KassaProcessingStats, 0),
	}

	var statsMutex sync.Mutex
	var wg sync.WaitGroup
	transactionDetailsMap := make(map[string]int)
	kassaDetailsMap := make(map[string]*KassaProcessingStats)
	getKassaStats := func(folder models.KassaFolder) *KassaProcessingStats {
		key := folder.KassaCode + "/" + folder.FolderName
		if detail, ok := kassaDetailsMap[key]; ok {
			return detail
		}
		detail := &KassaProcessingStats{
			KassaCode:    folder.KassaCode,
			FolderName:   folder.FolderName,
			SourceFolder: key,
			RequestPath:  folder.RequestPath,
			ResponsePath: folder.ResponsePath,
		}
		kassaDetailsMap[key] = detail
		return detail
	}

	logger.InfoContext(ctx, "Found kassa folders to process",
		"count", len(folders),
		"response_wait_delay", cfg.WaitDelayMinutes.String(),
		"lock_retry_delay", cfg.RetryDelay.String(),
		"event", "ftp_folders_found",
	)

	for _, folder := range folders {
		folder := folder
		wg.Add(1)
		go func() {
			defer wg.Done()
			folderStats := processFolderLoad(ctx, ftpClient, loader, cfg, date, folder, logger)

			statsMutex.Lock()
			defer statsMutex.Unlock()

			stats.FilesProcessed += folderStats.Detail.FilesProcessed
			stats.FilesSkipped += folderStats.Detail.FilesSkipped
			stats.FilesRecovered += folderStats.Detail.FilesRecovered
			stats.TransactionsLoaded += 0
			for _, count := range folderStats.TransactionDetails {
				stats.TransactionsLoaded += count
			}
			for stage, count := range folderStats.ErrorBreakdown {
				stats.ErrorBreakdown[stage] += count
				stats.Errors += count
			}
			for _, sample := range folderStats.ErrorSamples {
				if len(stats.ErrorSamples) >= maxErrorSamples {
					break
				}
				stats.ErrorSamples = append(stats.ErrorSamples, sample)
			}

			detail := getKassaStats(folder)
			*detail = folderStats.Detail
			for tableName, count := range folderStats.TransactionDetails {
				transactionDetailsMap[tableName] += count
			}
		}()
	}

	wg.Wait()

	// Преобразуем агрегированную статистику в слайс
	for tableName, count := range transactionDetailsMap {
		stats.TransactionDetails = append(stats.TransactionDetails, TransactionTypeStats{
			TableName: tableName,
			Count:     count,
		})
	}
	for _, detail := range kassaDetailsMap {
		stats.KassaDetails = append(stats.KassaDetails, *detail)
	}
	sort.Slice(stats.KassaDetails, func(i, j int) bool {
		return stats.KassaDetails[i].SourceFolder < stats.KassaDetails[j].SourceFolder
	})

	// Выводим финальную статистику
	logger.InfoContext(ctx, "Processing statistics",
		"log_kind", "loki_operational",
		"files_processed", stats.FilesProcessed,
		"files_skipped", stats.FilesSkipped,
		"files_recovered", stats.FilesRecovered,
		"errors", stats.Errors,
		"kassa_details", stats.KassaDetails,
		"error_breakdown", stats.ErrorBreakdown,
		"error_samples", stats.ErrorSamples,
		"event", "processing_statistics",
	)

	return stats, nil
}

func processFolderLoad(ctx context.Context, ftpClient ftp.FTPClient, loader fileLoader, cfg *models.Config, date string, folder models.KassaFolder, logger *slog.Logger) folderRunResult {
	sourceFolder := folder.KassaCode + "/" + folder.FolderName
	result := folderRunResult{
		Detail: KassaProcessingStats{
			KassaCode:    folder.KassaCode,
			FolderName:   folder.FolderName,
			SourceFolder: sourceFolder,
			Status:       "pending",
			RequestPath:  folder.RequestPath,
			ResponsePath: folder.ResponsePath,
		},
		ErrorBreakdown:     make(map[string]int),
		TransactionDetails: make(map[string]int),
	}

	recordError := func(stage, file, path string, err error) {
		result.ErrorBreakdown[stage]++
		result.Detail.FilesFailed++
		result.Detail.Status = stage
		result.Detail.LastIssueStage = stage
		if err != nil {
			result.Detail.LastIssueMessage = err.Error()
		}
		if len(result.ErrorSamples) >= maxErrorSamples {
			return
		}
		sample := PipelineIssueSample{Stage: stage}
		if file != "" {
			sample.File = file
		}
		if path != "" {
			sample.Path = path
		}
		if err != nil {
			sample.Error = err.Error()
		}
		result.ErrorSamples = append(result.ErrorSamples, sample)
	}

	releaseLock, lockWait, err := defaultFolderLocks.acquire(ctx, sourceFolder, cfg.RetryDelay, cfg.WaitDelayMinutes)
	if err != nil {
		recordError("folder_lock_error", "", folder.ResponsePath, err)
		result.Detail.LockWait = lockWait.String()
		return result
	}
	defer releaseLock()
	result.Detail.LockWait = lockWait.String()
	result.Detail.Status = "lock_acquired"

	deletedResponses, err := cleanupFolderPath(ctx, ftpClient, folder.ResponsePath, sourceFolder, "response", logger)
	if err != nil {
		recordError("response_cleanup_failed", "", folder.ResponsePath, err)
		return result
	}
	result.Detail.DeletedResponses = deletedResponses
	result.Detail.Status = "response_cleaned"

	remainingResponseFiles, err := ftpClient.ListFiles(folder.ResponsePath)
	if err != nil {
		recordError("response_preflight_failed", "", folder.ResponsePath, err)
		return result
	}
	if len(remainingResponseFiles) > 0 {
		recordError("response_preflight_failed", "", folder.ResponsePath, fmt.Errorf("response folder still contains %d files after cleanup", len(remainingResponseFiles)))
		return result
	}

	deletedRequests, err := cleanupFolderPath(ctx, ftpClient, folder.RequestPath, sourceFolder, "request", logger)
	if err != nil {
		recordError("request_cleanup_failed", "", folder.RequestPath, err)
		return result
	}
	result.Detail.DeletedRequests = deletedRequests
	result.Detail.Status = "request_cleaned"

	if err := ftpClient.SendRequestToKassa(folder, date); err != nil {
		recordError("request_send_failed", "", folder.RequestPath, err)
		return result
	}
	result.Detail.Status = "request_sent"

	if err := waitForResponses(ctx, cfg.WaitDelayMinutes); err != nil {
		recordError("response_wait_canceled", "", folder.ResponsePath, err)
		return result
	}
	result.Detail.Status = "waiting_response"

	allFiles, responseFiles, skippedFiles, err := listProcessableResponseFiles(ftpClient, folder.ResponsePath)
	if err != nil {
		recordError("response_list_failed", "", folder.ResponsePath, err)
		return result
	}

	result.Detail.FilesFound = len(allFiles)
	result.Detail.FilesSkipped = skippedFiles
	result.Detail.FilesQueued = len(responseFiles)

	if len(responseFiles) == 0 {
		recordError("no_response", "", folder.ResponsePath, fmt.Errorf("no processable response files found after wait"))
		return result
	}
	result.Detail.Status = "processing_response"

	for _, file := range responseFiles {
		outcome, err := processFile(ctx, ftpClient, loader, cfg, file.Name, folder, date, logger)
		if err != nil {
			recordError(stageForFileError(err), file.Name, folder.ResponsePath, err)
			continue
		}

		result.Detail.FilesProcessed++
		if outcome.Recovered {
			result.Detail.FilesRecovered++
		}
		for _, detail := range outcome.TransactionDetails {
			tableName, _ := detail["table_name"].(string)
			count, _ := detail["count"].(int)
			result.TransactionDetails[tableName] += count
		}
		logger.InfoContext(ctx, "Successfully processed file",
			"file", file.Name,
			"source_folder", sourceFolder,
			"transactions_loaded", outcome.LoadedTransactions,
			"recovered", outcome.Recovered,
			"event", "file_process_success",
		)
	}

	if len(result.ErrorBreakdown) == 0 {
		result.Detail.Status = "loaded"
	} else if result.Detail.FilesProcessed > 0 {
		result.Detail.Status = "partial"
	}

	return result
}

func cleanupFolderPath(ctx context.Context, ftpClient ftp.FTPClient, path, sourceFolder, folderKind string, logger *slog.Logger) (int, error) {
	files, err := ftpClient.ListFiles(path)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list folder before cleanup",
			"source_folder", sourceFolder,
			"path", path,
			"folder_kind", folderKind,
			"error", err.Error(),
			"event", "folder_cleanup_list_error",
		)
		return 0, err
	}
	if err := ftpClient.ClearDirectory(path); err != nil {
		logger.ErrorContext(ctx, "Failed to clear folder",
			"source_folder", sourceFolder,
			"path", path,
			"folder_kind", folderKind,
			"error", err.Error(),
			"event", "folder_cleanup_error",
		)
		return 0, err
	}
	logger.InfoContext(ctx, "Folder cleanup complete",
		"source_folder", sourceFolder,
		"path", path,
		"folder_kind", folderKind,
		"deleted_files", len(files),
		"event", "folder_cleanup_complete",
	)
	return len(files), nil
}

func listProcessableResponseFiles(ftpClient ftp.FTPClient, responsePath string) ([]*ftplib.Entry, []*ftplib.Entry, int, error) {
	allFiles, err := ftpClient.ListFiles(responsePath)
	if err != nil {
		return nil, nil, 0, err
	}
	responseFiles := ftp.FilterUnprocessedFiles(allFiles)
	responseFiles = ftp.FilterFilesByName(responseFiles, "SaveResult001.txt")
	return allFiles, responseFiles, len(allFiles) - len(responseFiles), nil
}

// processFile обрабатывает один файл из FTP
// Возвращает количество транзакций, детальную статистику и ошибку
func processFile(ctx context.Context, ftpClient ftp.FTPClient, loader fileLoader, cfg *models.Config, filename string, folder models.KassaFolder, requestedDate string, logger *slog.Logger) (fileProcessOutcome, error) {
	outcome := fileProcessOutcome{}
	store := newFileLifecycleStore(cfg.LocalDir)

	// Скачиваем файл с FTP
	// Используем уникальный путь для локального файла, включая информацию о папке,
	// чтобы избежать конфликтов при параллельной обработке файлов с одинаковыми именами из разных папок
	remotePath := folder.ResponsePath + "/" + filename
	logicalKey := remotePath + "|" + requestedDate
	// Создаем уникальный локальный путь: LocalDir/KassaCode/FolderName/filename
	localPath := fmt.Sprintf("%s/%s/%s/%s", cfg.LocalDir, folder.KassaCode, folder.FolderName, filename)

	err := ftpClient.DownloadFile(remotePath, localPath)
	if err != nil {
		return outcome, newStagedFileError("file_download_error", fmt.Errorf("failed to download file: %w", err))
	}
	defer func() {
		// Очищаем локальный файл
		if err := removeFile(localPath); err != nil {
			logger.WarnContext(ctx, "Failed to remove local file",
				"file", localPath,
				"error", err.Error(),
			)
		}
	}()

	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return outcome, newStagedFileError("file_stat_error", fmt.Errorf("failed to stat local file: %w", err))
	}
	if fileInfo.Size() == 0 {
		return outcome, newStagedFileError("empty_response", errors.New("response file is empty"))
	}

	sourceFolder := folder.KassaCode + "/" + folder.FolderName
	contentHash, err := hashLocalFile(localPath)
	if err != nil {
		return outcome, newStagedFileError("file_hash_error", fmt.Errorf("failed to hash file: %w", err))
	}

	existingState, err := store.Load(store.key(logicalKey, contentHash))
	if err != nil {
		return outcome, newStagedFileError("file_state_load_error", fmt.Errorf("failed to load file lifecycle state: %w", err))
	}
	latestState, err := store.LoadLatest(logicalKey)
	if err != nil {
		return outcome, newStagedFileError("file_state_load_error", fmt.Errorf("failed to load latest file lifecycle state: %w", err))
	}
	if existingState != nil {
		switch existingState.Stage {
		case fileLifecycleStagePendingFinalize:
			logger.InfoContext(ctx, "Recovering previously loaded file by finalizing FTP state",
				"file", filename,
				"remote_path", remotePath,
				"event", "file_recovery_finalize",
			)
			if err := finalizeFileLifecycle(ctx, ftpClient, store, existingState, logger); err != nil {
				return outcome, newStagedFileError("file_finalize_error", err)
			}
			outcome.Recovered = true
			return outcome, nil
		case fileLifecycleStageParseFailed:
			return outcome, newStagedFileError("file_quarantined", fmt.Errorf("file is quarantined after previous parse failure: %s", existingState.LastError))
		}
	}
	if latestState != nil && latestState.Stage == fileLifecycleStageCompleted && latestState.ContentHash == contentHash {
		logger.InfoContext(ctx, "File content already loaded previously, finalizing FTP state without reload",
			"file", filename,
			"remote_path", remotePath,
			"event", "file_duplicate_finalize",
		)
		if err := finalizeFileLifecycle(ctx, ftpClient, store, latestState, logger); err != nil {
			return outcome, newStagedFileError("file_finalize_error", err)
		}
		outcome.Recovered = true
		return outcome, nil
	}

	// Парсим файл
	transactions, header, err := parser.ParseFile(localPath, sourceFolder)
	if err != nil {
		record := newFileLifecycleRecord(store, logicalKey, remotePath, requestedDate, filename, sourceFolder, nil, contentHash, 0)
		if saveErr := store.Save(record.withStage(fileLifecycleStageParseFailed, err.Error())); saveErr != nil {
			return outcome, newStagedFileError("file_parse_error", fmt.Errorf("failed to parse file: %w (also failed to persist parse failure: %v)", err, saveErr))
		}
		return outcome, newStagedFileError("file_parse_error", fmt.Errorf("failed to parse file: %w", err))
	}

	// Выводим информацию о заголовке файла
	logger.DebugContext(ctx, "File header",
		"file", filename,
		"processed", header.Processed,
		"db_id", header.DBID,
		"report_number", header.ReportNum,
		"event", "file_header",
	)

	// Проверяем, обработан ли файл уже
	if header.Processed {
		logger.InfoContext(ctx, "File is already marked as processed in header, finalizing FTP state",
			"file", filename,
			"event", "file_already_processed",
		)
		record := newFileLifecycleRecord(store, logicalKey, remotePath, requestedDate, filename, sourceFolder, header, contentHash, 0)
		if err := finalizeFileLifecycle(ctx, ftpClient, store, record, logger); err != nil {
			return outcome, newStagedFileError("file_finalize_error", err)
		}
		outcome.Recovered = true
		return outcome, nil
	}

	// Подсчитываем количество транзакций перед загрузкой
	transactionCount := loader.GetTransactionCount(transactions)
	manifest, err := buildTransactionManifest(transactions)
	if err != nil {
		return outcome, newStagedFileError("file_manifest_error", fmt.Errorf("failed to build transaction manifest: %w", err))
	}
	record := newFileLifecycleRecord(store, logicalKey, remotePath, requestedDate, filename, sourceFolder, header, contentHash, transactionCount).withManifest(manifest)

	logger.InfoContext(ctx, "Found transactions in file",
		"file", filename,
		"transaction_count", transactionCount,
		"event", "transactions_found",
	)

	staleManifest := map[string][]int64(nil)
	if latestState != nil && latestState.Stage == fileLifecycleStageCompleted && latestState.ContentHash != contentHash {
		staleManifest = latestState.TransactionManifest
		logger.InfoContext(ctx, "Detected corrected reupload, reconciling stale rows before reload",
			"file", filename,
			"remote_path", remotePath,
			"previous_hash", latestState.ContentHash,
			"current_hash", contentHash,
			"event", "file_reupload_reconcile",
		)
	}

	// Загружаем данные в базу данных только если есть транзакции
	if transactionCount > 0 || len(staleManifest) > 0 {
		// Создаем отдельный контекст для загрузки данных с увеличенным таймаутом
		// Сохраняем родительский контекст для корректной propagation отмены
		loadCtx, loadCancel := context.WithTimeout(ctx, cfg.EffectivePipelineLoadTimeout())
		defer loadCancel()

		if err := loader.LoadFileDataWithReconcile(loadCtx, sourceFolder, staleManifest, transactions); err != nil {
			return outcome, newStagedFileError("file_load_error", fmt.Errorf("failed to load data: %w", err))
		}
		logger.InfoContext(ctx, "Successfully loaded transactions into database",
			"transaction_count", transactionCount,
			"event", "transactions_loaded",
		)
		outcome.LoadedTransactions = transactionCount
		outcome.TransactionDetails = loader.GetTransactionDetails(transactions)
	} else {
		logger.DebugContext(ctx, "No transactions to load (file contains only header)",
			"event", "no_transactions",
		)
	}

	if err := store.Save(record.withStage(fileLifecycleStagePendingFinalize, "")); err != nil {
		return outcome, newStagedFileError("file_state_save_error", fmt.Errorf("failed to persist file lifecycle state: %w", err))
	}

	// Отмечаем файл как обработанный только после успешной загрузки (или если файл пустой)
	if err := finalizeFileLifecycle(ctx, ftpClient, store, record.withStage(fileLifecycleStagePendingFinalize, ""), logger); err != nil {
		return outcome, newStagedFileError("file_finalize_error", err)
	}

	return outcome, nil
}

func buildTransactionManifest(transactions map[string]interface{}) (map[string][]int64, error) {
	manifest := make(map[string][]int64)
	for tableName, data := range transactions {
		rv := reflect.ValueOf(data)
		if !rv.IsValid() || rv.Kind() != reflect.Slice || rv.Len() == 0 {
			continue
		}
		ids := make([]int64, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			row := rv.Index(i)
			if row.Kind() == reflect.Pointer {
				row = row.Elem()
			}
			if row.Kind() != reflect.Struct {
				return nil, fmt.Errorf("table %s row %d is not struct", tableName, i)
			}
			field := row.FieldByName("TransactionIDUnique")
			if !field.IsValid() || field.Kind() != reflect.Int64 {
				return nil, fmt.Errorf("table %s row %d missing TransactionIDUnique", tableName, i)
			}
			ids = append(ids, field.Int())
		}
		manifest[tableName] = ids
	}
	return manifest, nil
}

// removeFile удаляет файл, игнорируя ошибки
func removeFile(path string) error {
	return os.Remove(path)
}

func waitForResponses(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func finalizeResult(result *PipelineResult) {
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()
	if result.Status == "" {
		result.Status = PipelineStatusFailed
	}
	result.Success = result.Status == PipelineStatusCompleted
}

type issueCollector struct {
	breakdown map[string]int
	samples   []PipelineIssueSample
}

func newIssueCollector() *issueCollector {
	return &issueCollector{breakdown: make(map[string]int)}
}

func (c *issueCollector) Record(stage string, file string, path string, err error) {
	c.breakdown[stage]++
	if len(c.samples) >= maxErrorSamples {
		return
	}
	sample := PipelineIssueSample{Stage: stage}
	if file != "" {
		sample.File = file
	}
	if path != "" {
		sample.Path = path
	}
	if err != nil {
		sample.Error = err.Error()
	}
	c.samples = append(c.samples, sample)
}

func (c *issueCollector) Merge(stats *ProcessingStats) {
	if stats == nil {
		return
	}
	for stage, count := range stats.ErrorBreakdown {
		c.breakdown[stage] += count
	}
	for _, sample := range stats.ErrorSamples {
		if len(c.samples) >= maxErrorSamples {
			break
		}
		c.samples = append(c.samples, sample)
	}
}

func (c *issueCollector) Total() int {
	total := 0
	for _, count := range c.breakdown {
		total += count
	}
	return total
}

func (c *issueCollector) CloneBreakdown() map[string]int {
	if len(c.breakdown) == 0 {
		return nil
	}
	cloned := make(map[string]int, len(c.breakdown))
	for stage, count := range c.breakdown {
		cloned[stage] = count
	}
	return cloned
}

func (c *issueCollector) CloneSamples() []PipelineIssueSample {
	if len(c.samples) == 0 {
		return nil
	}
	cloned := make([]PipelineIssueSample, len(c.samples))
	copy(cloned, c.samples)
	return cloned
}

func (c *issueCollector) Summary() string {
	if c.Total() == 0 {
		return ""
	}
	if len(c.samples) == 0 {
		return fmt.Sprintf("ETL completed partially with %d errors", c.Total())
	}
	return fmt.Sprintf("ETL completed partially with %d errors; first issue at %s: %s", c.Total(), c.samples[0].Stage, c.samples[0].Error)
}
