package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/repository"
)

const (
	exportWriterBufferSize = 32 * 1024
	exportFlushInterval    = 128
	maxSkippedIDSamples    = 20
)

type exportWriteStats struct {
	Written             int
	Skipped             int
	SkippedIDSamples    []int64
	SkippedIDsTruncated bool
}

type downloadResult struct {
	StatusCode          int
	Outcome             string
	RowsRetrieved       int
	RowsWritten         int
	RowsSkipped         int
	SkippedIDsTruncated bool
}

// processDownloadRequest обрабатывает синхронную выгрузку данных.
func (s *Server) processDownloadRequest(ctx context.Context, log *logger.Logger, w http.ResponseWriter, sourceFolder string, date string) downloadResult {
	result := downloadResult{StatusCode: http.StatusOK, Outcome: "exported"}
	log.InfoContext(ctx, "Processing download operation",
		"log_kind", "loki_operational",
		"source_folder", sourceFolder,
		"date", date,
		"event", "download_processing_start",
	)

	database, err := db.NewPool(s.config)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database",
			"error", err.Error(),
			"event", "db_connection_error",
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		result.StatusCode = http.StatusInternalServerError
		result.Outcome = "db_connection_error"
		return result
	}
	defer database.Close()

	loader := repository.NewLoader(database)

	allTransactions, err := loader.GetAllTransactionsBySourceFolderAndDate(ctx, sourceFolder, date)
	if err != nil {
		log.ErrorContext(ctx, "Failed to get all transactions",
			"error", err.Error(),
			"event", "query_error",
		)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		result.StatusCode = http.StatusInternalServerError
		result.Outcome = "query_error"
		return result
	}
	result.RowsRetrieved = len(allTransactions)

	if len(allTransactions) == 0 {
		log.InfoContext(ctx, "No transactions found",
			"source_folder", sourceFolder,
			"date", date,
			"event", "no_transactions_found",
		)
		http.Error(w, "No transactions found for the specified source_folder and date", http.StatusNotFound)
		result.StatusCode = http.StatusNotFound
		result.Outcome = "not_found"
		return result
	}

	log.InfoContext(ctx, "All transactions retrieved",
		"count", len(allTransactions),
		"source_folder", sourceFolder,
		"event", "all_transactions_retrieved",
	)

	filename := fmt.Sprintf("kassa_%s_%s.txt", strings.ReplaceAll(sourceFolder, "/", "_"), date)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	var flusher http.Flusher
	if responseFlusher, ok := w.(http.Flusher); ok {
		flusher = responseFlusher
	}

	stats, err := streamExportTXT(w, flusher, allTransactions)
	if err != nil {
		log.ErrorContext(ctx, "Failed to send TXT data",
			"error", err.Error(),
			"event", "txt_send_error",
		)
		result.StatusCode = http.StatusInternalServerError
		result.Outcome = "stream_write_error"
		return result
	}
	result.RowsWritten = stats.Written
	result.RowsSkipped = stats.Skipped
	result.SkippedIDsTruncated = stats.SkippedIDsTruncated

	log.InfoContext(ctx, "Transactions written to file",
		"log_kind", "loki_operational",
		"written", stats.Written,
		"skipped", stats.Skipped,
		"skipped_id_samples", stats.SkippedIDSamples,
		"skipped_ids_truncated", stats.SkippedIDsTruncated,
		"total_retrieved", len(allTransactions),
		"event", "transactions_write_stats",
	)

	log.InfoContext(ctx, "TXT file sent successfully",
		"log_kind", "loki_operational",
		"rows", stats.Written,
		"event", "txt_sent",
	)
	log.InfoContext(ctx, "Download export summary",
		"log_kind", "loki_operational",
		"source_folder", sourceFolder,
		"date", date,
		"rows_retrieved", result.RowsRetrieved,
		"rows_written", result.RowsWritten,
		"rows_skipped", result.RowsSkipped,
		"skipped_ids_truncated", result.SkippedIDsTruncated,
		"event", "download_export_summary",
	)
	return result
}

func streamExportTXT(w io.Writer, flusher http.Flusher, transactions []repository.ExportRow) (exportWriteStats, error) {
	stats := exportWriteStats{SkippedIDSamples: make([]int64, 0, maxSkippedIDSamples)}
	writer := bufio.NewWriterSize(w, exportWriterBufferSize)

	// #nosec G705 -- export content is returned as plain text, not rendered as HTML.
	if _, err := fmt.Fprintf(writer, "#\n1\n%d\n", len(transactions)); err != nil {
		return stats, fmt.Errorf("write export header: %w", err)
	}

	for i, tx := range transactions {
		if tx.RawLine == "" {
			stats.Skipped++
			if len(stats.SkippedIDSamples) < maxSkippedIDSamples {
				stats.SkippedIDSamples = append(stats.SkippedIDSamples, tx.TransactionIDUnique)
			} else {
				stats.SkippedIDsTruncated = true
			}
			continue
		}

		if _, err := writer.WriteString(tx.RawLine); err != nil {
			return stats, fmt.Errorf("write export line: %w", err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return stats, fmt.Errorf("write export newline: %w", err)
		}
		stats.Written++

		if flusher != nil && (i+1)%exportFlushInterval == 0 {
			if err := writer.Flush(); err != nil {
				return stats, fmt.Errorf("flush export chunk: %w", err)
			}
			flusher.Flush()
		}
	}

	if err := writer.Flush(); err != nil {
		return stats, fmt.Errorf("flush export writer: %w", err)
	}
	if flusher != nil {
		flusher.Flush()
	}

	return stats, nil
}
