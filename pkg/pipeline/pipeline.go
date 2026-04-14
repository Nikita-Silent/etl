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
	"github.com/user/go-frontol-loader/pkg/workers"
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

var (
	processFilesFromFTPFunc = processFilesFromFTP
	waitForResponsesFunc    = waitForResponses
)

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
	ResponsePath     string `json:"response_path"`
	FilesFound       int    `json:"files_found"`
	FilesQueued      int    `json:"files_queued"`
	FilesProcessed   int    `json:"files_processed"`
	FilesSkipped     int    `json:"files_skipped"`
	FilesRecovered   int    `json:"files_recovered,omitempty"`
	FilesFailed      int    `json:"files_failed,omitempty"`
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

	// Шаг 1: Очистка только request папок (response папки не очищаем, чтобы не удалить существующие файлы)
	logger.InfoContext(ctx, "Step 1: Clearing request folders only (keeping response files)",
		"event", "etl_step_1",
	)
	if err := ftpClient.ClearAllKassaRequestFolders(); err != nil {
		issues.Record("clear_request_folders", "", "", err)
		logger.WarnContext(ctx, "Failed to clear some request folders",
			"error", err.Error(),
			"event", "etl_step_1_warning",
		)
		// Не прерываем выполнение, продолжаем
	} else {
		logger.InfoContext(ctx, "Step 1: All request folders cleared successfully",
			"event", "etl_step_1_complete",
		)
	}

	// Шаг 2: Отправка request.txt файлов во все кассы
	logger.InfoContext(ctx, "Step 2: Sending request.txt files to all kassas",
		"date", date,
		"event", "etl_step_2",
	)
	if err := ftpClient.SendRequestsToAllKassasWithDate(date); err != nil {
		issues.Record("send_requests", "", "", err)
		logger.WarnContext(ctx, "Failed to send some requests",
			"error", err.Error(),
			"event", "etl_step_2_warning",
		)
		// Не прерываем выполнение, продолжаем
	} else {
		logger.InfoContext(ctx, "Step 2: Request files sent successfully to all kassas",
			"event", "etl_step_2_complete",
		)
	}

	// Шаг 3: Ожидание генерации ответов
	logger.InfoContext(ctx, "Step 3: Waiting for responses to be generated",
		"wait_delay", cfg.WaitDelayMinutes.String(),
		"event", "etl_step_3",
	)
	if err := waitForResponsesFunc(ctx, cfg.WaitDelayMinutes); err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed while waiting for responses: %v", err)
		result.Errors = issues.Total()
		result.ErrorBreakdown = issues.CloneBreakdown()
		result.ErrorSamples = issues.CloneSamples()
		logger.ErrorContext(ctx, "Failed while waiting for responses",
			"error", err.Error(),
			"event", "etl_step_3_error",
		)
		return result, err
	}

	// Шаг 4: Обработка файлов из FTP
	logger.InfoContext(ctx, "Step 4: Processing files from FTP",
		"event", "etl_step_4",
	)
	logger.DebugContext(ctx, "FTP configuration",
		"ftp_request_dir", cfg.FTPRequestDir,
		"ftp_response_dir", cfg.FTPResponseDir,
		"kassa_structure", fmt.Sprintf("%v", cfg.KassaStructure),
	)

	// Очищаем старые .processed файлы перед обработкой новых
	logger.InfoContext(ctx, "Clearing old .processed files from response folders",
		"event", "etl_clear_processed",
	)
	if err := ftpClient.ClearAllKassaResponseProcessedFiles(); err != nil {
		issues.Record("clear_processed_files", "", "", err)
		logger.WarnContext(ctx, "Failed to clear some processed files",
			"error", err.Error(),
			"event", "etl_clear_processed_warning",
		)
		// Не прерываем выполнение, продолжаем обработку
	}

	stats, err := processFilesFromFTPFunc(ctx, ftpClient, loader, cfg, logger)
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
			"files_found", detail.FilesFound,
			"files_queued", detail.FilesQueued,
			"files_processed", detail.FilesProcessed,
			"files_skipped", detail.FilesSkipped,
			"files_recovered", detail.FilesRecovered,
			"files_failed", detail.FilesFailed,
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

// fileTask представляет задачу на обработку файла
type fileTask struct {
	folder models.KassaFolder
	file   *ftplib.Entry
}

// processFilesFromFTP обрабатывает все файлы из FTP папок асинхронно
func processFilesFromFTP(ctx context.Context, ftpClient ftp.FTPClient, loader fileLoader, cfg *models.Config, logger *slog.Logger) (*ProcessingStats, error) {
	// Получаем все папки касс
	folders := ftp.GetAllKassaFolders(cfg)

	stats := &ProcessingStats{
		TransactionDetails: make([]TransactionTypeStats, 0),
		ErrorBreakdown:     make(map[string]int),
		ErrorSamples:       make([]PipelineIssueSample, 0, maxErrorSamples),
		KassaDetails:       make([]KassaProcessingStats, 0),
	}

	// Мьютекс для защиты общих структур данных
	var statsMutex sync.Mutex
	var errorMutex sync.Mutex

	// Map для агрегации статистики по типам транзакций
	transactionDetailsMap := make(map[string]int)
	kassaDetailsMap := make(map[string]*KassaProcessingStats)
	var totalFiles int
	getKassaStats := func(folder models.KassaFolder) *KassaProcessingStats {
		key := folder.KassaCode + "/" + folder.FolderName
		if detail, ok := kassaDetailsMap[key]; ok {
			return detail
		}
		detail := &KassaProcessingStats{
			KassaCode:    folder.KassaCode,
			FolderName:   folder.FolderName,
			SourceFolder: key,
			ResponsePath: folder.ResponsePath,
		}
		kassaDetailsMap[key] = detail
		return detail
	}
	recordError := func(stage string, file string, path string, err error) {
		errorMutex.Lock()
		defer errorMutex.Unlock()
		stats.ErrorBreakdown[stage]++
		if len(stats.ErrorSamples) >= maxErrorSamples {
			return
		}
		sample := PipelineIssueSample{
			Stage: stage,
		}
		if file != "" {
			sample.File = file
		}
		if path != "" {
			sample.Path = path
		}
		if err != nil {
			sample.Error = err.Error()
		}
		stats.ErrorSamples = append(stats.ErrorSamples, sample)
	}

	logger.InfoContext(ctx, "Found kassa folders to process",
		"count", len(folders),
		"event", "ftp_folders_found",
	)

	// Собираем все задачи на обработку файлов
	var tasks []fileTask

	for _, folder := range folders {
		logger.InfoContext(ctx, "Scanning kassa folder",
			"kassa_code", folder.KassaCode,
			"folder_name", folder.FolderName,
			"request_path", folder.RequestPath,
			"response_path", folder.ResponsePath,
			"event", "ftp_scanning_kassa",
		)

		// Список файлов в папке ответов
		allFiles, err := ftpClient.ListFiles(folder.ResponsePath)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to list files",
				"path", folder.ResponsePath,
				"error", err.Error(),
				"event", "ftp_list_error",
			)
			statsMutex.Lock()
			stats.Errors++
			detail := getKassaStats(folder)
			detail.FilesFailed++
			detail.LastIssueStage = "ftp_list_error"
			detail.LastIssueMessage = err.Error()
			statsMutex.Unlock()
			recordError("ftp_list_error", "", folder.ResponsePath, err)
			continue
		}

		totalFiles += len(allFiles)

		// Фильтруем необработанные файлы (O(n) вместо O(n²))
		// Это делает один проход по списку вместо множественных FTP запросов
		unprocessedFiles := ftp.FilterUnprocessedFiles(allFiles)

		// Исключаем специальные файлы
		unprocessedFiles = ftp.FilterFilesByName(unprocessedFiles, "SaveResult001.txt")

		// Подсчитываем пропущенные файлы
		skippedCount := len(allFiles) - len(unprocessedFiles)
		if skippedCount > 0 {
			statsMutex.Lock()
			stats.FilesSkipped += skippedCount
			getKassaStats(folder).FilesSkipped += skippedCount
			statsMutex.Unlock()
		}

		statsMutex.Lock()
		detail := getKassaStats(folder)
		detail.FilesFound += len(allFiles)
		detail.FilesQueued += len(unprocessedFiles)
		statsMutex.Unlock()

		logger.InfoContext(ctx, "Found files in response folder",
			"path", folder.ResponsePath,
			"total_files", len(allFiles),
			"unprocessed_files", len(unprocessedFiles),
			"skipped_files", skippedCount,
			"event", "ftp_files_found",
		)

		if len(unprocessedFiles) > 0 {
			fileNames := make([]string, len(unprocessedFiles))
			for i, f := range unprocessedFiles {
				fileNames[i] = f.Name
			}
			logger.DebugContext(ctx, "Unprocessed files",
				"files", fileNames,
			)
		} else if len(allFiles) == 0 {
			logger.WarnContext(ctx, "No files found in response folder",
				"path", folder.ResponsePath,
				"event", "ftp_no_files",
			)
		}

		// Добавляем необработанные файлы в список задач
		for _, file := range unprocessedFiles {
			tasks = append(tasks, fileTask{
				folder: folder,
				file:   file,
			})
		}
	}

	logger.InfoContext(ctx, "Starting asynchronous file processing with worker pool",
		"total_tasks", len(tasks),
		"total_files_found", totalFiles,
		"worker_pool_size", cfg.WorkerPoolSize,
		"event", "async_processing_start",
	)

	// Создаем worker pool для ограничения количества параллельных операций
	pool := workers.NewPool(cfg.WorkerPoolSize)

	// Обрабатываем все файлы асинхронно через worker pool
	for _, task := range tasks {
		t := task // Capture loop variable
		err := pool.Submit(ctx, func() error {
			// Обрабатываем файл
			outcome, err := processFile(ctx, ftpClient, loader, cfg, t.file.Name, t.folder, logger)

			statsMutex.Lock()
			defer statsMutex.Unlock()

			if err != nil {
				logger.ErrorContext(ctx, "Failed to process file",
					"file", t.file.Name,
					"error", err.Error(),
					"event", "file_process_error",
				)
				stats.Errors++
				detail := getKassaStats(t.folder)
				detail.FilesFailed++
				detail.LastIssueStage = stageForFileError(err)
				detail.LastIssueMessage = err.Error()
				recordError(stageForFileError(err), t.file.Name, "", err)
				return err
			}

			stats.FilesProcessed++
			detail := getKassaStats(t.folder)
			detail.FilesProcessed++
			stats.TransactionsLoaded += outcome.LoadedTransactions
			if outcome.Recovered {
				stats.FilesRecovered++
				detail.FilesRecovered++
			}

			// Агрегируем детальную статистику
			for _, detail := range outcome.TransactionDetails {
				tableName, _ := detail["table_name"].(string)
				count, _ := detail["count"].(int)
				transactionDetailsMap[tableName] += count
			}

			logger.InfoContext(ctx, "Successfully processed file",
				"file", t.file.Name,
				"transactions_loaded", outcome.LoadedTransactions,
				"recovered", outcome.Recovered,
				"event", "file_process_success",
			)
			return nil
		})

		if err != nil {
			logger.ErrorContext(ctx, "Failed to submit file processing task",
				"file", t.file.Name,
				"error", err.Error(),
				"event", "worker_submit_error",
			)
			statsMutex.Lock()
			stats.Errors++
			statsMutex.Unlock()
			recordError("worker_submit_error", t.file.Name, "", err)
		}
	}

	// Ждем завершения всех workers
	pool.Wait()

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
		"total_files_found", totalFiles,
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

// processFile обрабатывает один файл из FTP
// Возвращает количество транзакций, детальную статистику и ошибку
func processFile(ctx context.Context, ftpClient ftp.FTPClient, loader fileLoader, cfg *models.Config, filename string, folder models.KassaFolder, logger *slog.Logger) (fileProcessOutcome, error) {
	outcome := fileProcessOutcome{}
	store := newFileLifecycleStore(cfg.LocalDir)

	// Скачиваем файл с FTP
	// Используем уникальный путь для локального файла, включая информацию о папке,
	// чтобы избежать конфликтов при параллельной обработке файлов с одинаковыми именами из разных папок
	remotePath := folder.ResponsePath + "/" + filename
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

	sourceFolder := folder.KassaCode + "/" + folder.FolderName
	contentHash, err := hashLocalFile(localPath)
	if err != nil {
		return outcome, newStagedFileError("file_hash_error", fmt.Errorf("failed to hash file: %w", err))
	}

	existingState, err := store.Load(store.key(remotePath, contentHash))
	if err != nil {
		return outcome, newStagedFileError("file_state_load_error", fmt.Errorf("failed to load file lifecycle state: %w", err))
	}
	latestState, err := store.LoadLatestByRemotePath(remotePath)
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
		record := newFileLifecycleRecord(store, remotePath, filename, sourceFolder, nil, contentHash, 0)
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
		record := newFileLifecycleRecord(store, remotePath, filename, sourceFolder, header, contentHash, 0)
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
	record := newFileLifecycleRecord(store, remotePath, filename, sourceFolder, header, contentHash, transactionCount).withManifest(manifest)

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
