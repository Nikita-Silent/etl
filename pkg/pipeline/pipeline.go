package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

// TransactionTypeStats представляет статистику по типу транзакций
type TransactionTypeStats struct {
	TableName string `json:"table_name"`
	Count     int    `json:"count"`
}

// PipelineResult содержит результат выполнения ETL-конвейера
type PipelineResult struct {
	StartTime          time.Time              `json:"start_time"`
	EndTime            time.Time              `json:"end_time"`
	Duration           string                 `json:"duration"`
	Date               string                 `json:"date"`
	FilesProcessed     int                    `json:"files_processed"`
	FilesSkipped       int                    `json:"files_skipped"`
	TransactionsLoaded int                    `json:"transactions_loaded"`
	Errors             int                    `json:"errors"`
	Success            bool                   `json:"success"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	TransactionDetails []TransactionTypeStats `json:"transaction_details,omitempty"` // Детальная информация по типам транзакций
}

// Run выполняет полный ETL-конвейер для указанной даты
func Run(ctx context.Context, logger *slog.Logger, cfg *models.Config, date string) (*PipelineResult, error) {
	result := &PipelineResult{
		StartTime: time.Now(),
		Date:      date,
		Success:   false,
	}

	logger.InfoContext(ctx, "Starting ETL pipeline",
		"date", date,
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

	// Шаг 1: Очистка только request папок (response папки не очищаем, чтобы не удалить существующие файлы)
	logger.InfoContext(ctx, "Step 1: Clearing request folders only (keeping response files)",
		"event", "etl_step_1",
	)
	if err := ftpClient.ClearAllKassaRequestFolders(); err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to clear some request folders: %v", err)
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
		result.ErrorMessage = fmt.Sprintf("Failed to send some requests: %v", err)
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
	time.Sleep(cfg.WaitDelayMinutes)

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
		logger.WarnContext(ctx, "Failed to clear some processed files",
			"error", err.Error(),
			"event", "etl_clear_processed_warning",
		)
		// Не прерываем выполнение, продолжаем обработку
	}

	stats, err := processFilesFromFTP(ctx, ftpClient, loader, cfg, logger)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to process files from FTP: %v", err)
		logger.ErrorContext(ctx, "Failed to process files from FTP",
			"error", err.Error(),
			"event", "etl_step_4_error",
		)
		return result, err
	}

	// Обновляем результат
	result.FilesProcessed = stats.FilesProcessed
	result.FilesSkipped = stats.FilesSkipped
	result.TransactionsLoaded = stats.TransactionsLoaded
	result.Errors = stats.Errors
	result.TransactionDetails = stats.TransactionDetails
	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	logger.InfoContext(ctx, "ETL pipeline completed successfully",
		"date", date,
		"files_processed", result.FilesProcessed,
		"files_skipped", result.FilesSkipped,
		"transactions_loaded", result.TransactionsLoaded,
		"errors", result.Errors,
		"event", "etl_complete",
	)

	return result, nil
}

// ProcessingStats содержит статистику обработки файлов
type ProcessingStats struct {
	FilesProcessed     int
	FilesSkipped       int
	TransactionsLoaded int
	Errors             int
	TransactionDetails []TransactionTypeStats // Детальная статистика по типам транзакций
}

// fileTask представляет задачу на обработку файла
type fileTask struct {
	folder models.KassaFolder
	file   *ftplib.Entry
}

// processFilesFromFTP обрабатывает все файлы из FTP папок асинхронно
func processFilesFromFTP(ctx context.Context, ftpClient ftp.FTPClient, loader *repository.Loader, cfg *models.Config, logger *slog.Logger) (*ProcessingStats, error) {
	// Получаем все папки касс
	folders := ftp.GetAllKassaFolders(cfg)

	stats := &ProcessingStats{
		TransactionDetails: make([]TransactionTypeStats, 0),
	}

	// Мьютекс для защиты общих структур данных
	var statsMutex sync.Mutex
	var errorMutex sync.Mutex

	// Map для агрегации статистики по типам транзакций
	transactionDetailsMap := make(map[string]int)
	var totalFiles int
	errorBreakdown := make(map[string]int)
	errorSamples := make([]map[string]string, 0, 5)
	const maxErrorSamples = 5

	recordError := func(stage string, file string, path string, err error) {
		errorMutex.Lock()
		defer errorMutex.Unlock()
		errorBreakdown[stage]++
		if len(errorSamples) >= maxErrorSamples {
			return
		}
		sample := map[string]string{
			"stage": stage,
		}
		if file != "" {
			sample["file"] = file
		}
		if path != "" {
			sample["path"] = path
		}
		if err != nil {
			sample["error"] = err.Error()
		}
		errorSamples = append(errorSamples, sample)
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
			statsMutex.Unlock()
		}

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
			transactionCount, fileDetails, err := processFile(ctx, ftpClient, loader, cfg, t.file.Name, t.folder, logger)

			statsMutex.Lock()
			defer statsMutex.Unlock()

			if err != nil {
				logger.ErrorContext(ctx, "Failed to process file",
					"file", t.file.Name,
					"error", err.Error(),
					"event", "file_process_error",
				)
				stats.Errors++
				recordError("file_process_error", t.file.Name, "", err)
				return err
			}

			stats.FilesProcessed++
			stats.TransactionsLoaded += transactionCount

			// Агрегируем детальную статистику
			for _, detail := range fileDetails {
				tableName, _ := detail["table_name"].(string)
				count, _ := detail["count"].(int)
				transactionDetailsMap[tableName] += count
			}

			logger.InfoContext(ctx, "Successfully processed file",
				"file", t.file.Name,
				"transactions_loaded", transactionCount,
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

	// Выводим финальную статистику
	logger.InfoContext(ctx, "Processing statistics",
		"total_files_found", totalFiles,
		"files_processed", stats.FilesProcessed,
		"files_skipped", stats.FilesSkipped,
		"errors", stats.Errors,
		"error_breakdown", errorBreakdown,
		"error_samples", errorSamples,
		"event", "processing_statistics",
	)

	return stats, nil
}

// processFile обрабатывает один файл из FTP
// Возвращает количество транзакций, детальную статистику и ошибку
func processFile(ctx context.Context, ftpClient ftp.FTPClient, loader *repository.Loader, cfg *models.Config, filename string, folder models.KassaFolder, logger *slog.Logger) (int, []map[string]interface{}, error) {
	// Скачиваем файл с FTP
	// Используем уникальный путь для локального файла, включая информацию о папке,
	// чтобы избежать конфликтов при параллельной обработке файлов с одинаковыми именами из разных папок
	remotePath := folder.ResponsePath + "/" + filename
	// Создаем уникальный локальный путь: LocalDir/KassaCode/FolderName/filename
	localPath := fmt.Sprintf("%s/%s/%s/%s", cfg.LocalDir, folder.KassaCode, folder.FolderName, filename)

	err := ftpClient.DownloadFile(remotePath, localPath)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to download file: %w", err)
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

	// Парсим файл
	sourceFolder := folder.KassaCode + "/" + folder.FolderName
	transactions, header, err := parser.ParseFile(localPath, sourceFolder)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse file: %w", err)
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
		logger.DebugContext(ctx, "File is already processed, skipping",
			"file", filename,
			"event", "file_already_processed",
		)
		return 0, nil, nil
	}

	// Подсчитываем количество транзакций перед загрузкой
	transactionCount := loader.GetTransactionCount(transactions)
	logger.InfoContext(ctx, "Found transactions in file",
		"file", filename,
		"transaction_count", transactionCount,
		"event", "transactions_found",
	)

	// Загружаем данные в базу данных только если есть транзакции
	if transactionCount > 0 {
		// Создаем отдельный контекст для загрузки данных с увеличенным таймаутом
		// Сохраняем родительский контекст для корректной propagation отмены
		loadCtx, loadCancel := context.WithTimeout(ctx, 1*time.Hour)
		defer loadCancel()

		if err := loader.LoadFileData(loadCtx, transactions); err != nil {
			return 0, nil, fmt.Errorf("failed to load data: %w", err)
		}
		logger.InfoContext(ctx, "Successfully loaded transactions into database",
			"transaction_count", transactionCount,
			"event", "transactions_loaded",
		)
	} else {
		logger.DebugContext(ctx, "No transactions to load (file contains only header)",
			"event", "no_transactions",
		)
	}

	// Отмечаем файл как обработанный только после успешной загрузки (или если файл пустой)
	err = ftpClient.MarkFileAsProcessed(remotePath)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to mark file as processed",
			"file", remotePath,
			"error", err.Error(),
			"event", "mark_processed_error",
		)
		return transactionCount, nil, fmt.Errorf("failed to mark file as processed: %w", err)
	}
	logger.DebugContext(ctx, "File marked as processed",
		"file", remotePath,
		"event", "file_marked_processed",
	)

	// Получаем детальную статистику по типам транзакций
	transactionDetails := loader.GetTransactionDetails(transactions)

	return transactionCount, transactionDetails, nil
}

// removeFile удаляет файл, игнорируя ошибки
func removeFile(path string) error {
	return os.Remove(path)
}
