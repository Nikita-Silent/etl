package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/user/go-frontol-loader/pkg/logger"
	"github.com/user/go-frontol-loader/pkg/pipeline"
)

var (
	runPipelineFunc            = pipeline.Run
	webhookTimeoutDurationFunc = func(minutes int) time.Duration {
		return time.Duration(minutes) * time.Minute
	}
)

// TransactionTypeStats представляет статистику по типу транзакций.
type TransactionTypeStats struct {
	TableName string `json:"table_name"`
	Count     int    `json:"count"`
}

// WebhookReport представляет отчет о выполнении ETL.
type WebhookReport struct {
	RequestID          string                          `json:"request_id"`
	Date               string                          `json:"date"`
	Status             string                          `json:"status"`
	StartTime          time.Time                       `json:"start_time"`
	EndTime            time.Time                       `json:"end_time"`
	Duration           string                          `json:"duration"`
	FilesProcessed     int                             `json:"files_processed"`
	FilesSkipped       int                             `json:"files_skipped"`
	FilesRecovered     int                             `json:"files_recovered,omitempty"`
	TransactionsLoaded int                             `json:"transactions_loaded"`
	Errors             int                             `json:"errors"`
	Success            bool                            `json:"success"`
	ErrorMessage       string                          `json:"error_message,omitempty"`
	ErrorBreakdown     map[string]int                  `json:"error_breakdown,omitempty"`
	ErrorSamples       []pipeline.PipelineIssueSample  `json:"error_samples,omitempty"`
	KassaDetails       []pipeline.KassaProcessingStats `json:"kassa_details,omitempty"`
	TransactionDetails []TransactionTypeStats          `json:"transaction_details,omitempty"`
}

// runETLPipeline запускает ETL pipeline и отправляет отчет.
func (s *Server) runETLPipeline(requestID, date string, log *logger.Logger) {
	ctx := context.Background()
	startTime := time.Now()
	report := &WebhookReport{
		RequestID: requestID,
		Date:      date,
		StartTime: startTime,
		Status:    "processing",
	}

	log.InfoContext(ctx, "Starting ETL pipeline",
		"log_kind", "loki_operational",
		"date", date,
		"event", "etl_pipeline_start",
	)

	pipelineDone := make(chan bool, 1)
	reportReady := make(chan *WebhookReport, 1)
	var reportMutex sync.Mutex
	var timeoutReportSent bool
	var finalReportSent bool

	go func() {
		defer func() {
			pipelineDone <- true
		}()

		result, err := runPipelineFunc(ctx, log.Logger, s.config, date)

		reportMutex.Lock()
		if err != nil {
			report.Status = "failed"
			report.ErrorMessage = err.Error()
			report.Success = false
			log.ErrorContext(ctx, "ETL pipeline failed",
				"log_kind", "loki_operational",
				"error", err.Error(),
				"event", "etl_pipeline_failed",
			)
		} else {
			report.Status = string(result.Status)
			report.Success = result.Success
			report.FilesProcessed = result.FilesProcessed
			report.FilesSkipped = result.FilesSkipped
			report.FilesRecovered = result.FilesRecovered
			report.TransactionsLoaded = result.TransactionsLoaded
			report.Errors = result.Errors
			report.ErrorMessage = result.ErrorMessage
			report.ErrorBreakdown = result.ErrorBreakdown
			report.ErrorSamples = result.ErrorSamples
			report.KassaDetails = result.KassaDetails
			report.TransactionDetails = make([]TransactionTypeStats, 0, len(result.TransactionDetails))
			for _, detail := range result.TransactionDetails {
				report.TransactionDetails = append(report.TransactionDetails, TransactionTypeStats{
					TableName: detail.TableName,
					Count:     detail.Count,
				})
			}
			logLevel := log.InfoContext
			if result.Status == pipeline.PipelineStatusPartial {
				logLevel = log.WarnContext
			}
			logLevel(ctx, "ETL pipeline completed",
				"log_kind", "loki_operational",
				"status", result.Status,
				"files_processed", result.FilesProcessed,
				"files_recovered", result.FilesRecovered,
				"transactions_loaded", result.TransactionsLoaded,
				"errors", result.Errors,
				"event", "etl_pipeline_completed",
			)
		}
		reportMutex.Unlock()

		reportReady <- report
	}()

	sendReport := func(r *WebhookReport, final bool) {
		reportMutex.Lock()
		if final {
			if finalReportSent {
				reportMutex.Unlock()
				return
			}
			finalReportSent = true
		} else {
			if timeoutReportSent {
				reportMutex.Unlock()
				return
			}
			timeoutReportSent = true
		}
		r.EndTime = time.Now()
		r.Duration = r.EndTime.Sub(r.StartTime).String()
		webhookConfigured := s.config.WebhookReportURL != ""
		reportMutex.Unlock()

		if s.config.WebhookReportURL != "" {
			log.InfoContext(ctx, "Sending webhook report",
				"log_kind", "loki_operational",
				"request_id", requestID,
				"date", date,
				"status", r.Status,
				"final", final,
				"event", "webhook_report_sending",
			)
			s.sendWebhookReport(r)
			log.InfoContext(ctx, "Webhook report sent",
				"log_kind", "loki_operational",
				"request_id", requestID,
				"date", date,
				"status", r.Status,
				"final", final,
				"event", "webhook_report_sent",
			)
		} else if !webhookConfigured {
			log.InfoContext(ctx, "Webhook report URL not configured, skipping report",
				"log_kind", "loki_operational",
				"request_id", requestID,
				"date", date,
				"final", final,
				"event", "webhook_report_skipped",
			)
		}
	}

	if s.config.WebhookTimeoutMinutes == 0 {
		log.InfoContext(ctx, "Waiting for pipeline completion (no timeout configured)",
			"log_kind", "loki_operational",
			"request_id", requestID,
			"date", date,
			"event", "waiting_pipeline_completion",
		)
		<-pipelineDone
		log.InfoContext(ctx, "Pipeline execution finished, waiting for report",
			"log_kind", "loki_operational",
			"request_id", requestID,
			"date", date,
			"event", "pipeline_done_waiting_report",
		)
		select {
		case r := <-reportReady:
			sendReport(r, true)
		case <-time.After(s.config.EffectiveWebhookReportResultWaitTimeout()):
			log.WarnContext(ctx, "Timeout waiting for report",
				"log_kind", "loki_operational",
				"request_id", requestID,
				"date", date,
				"event", "report_timeout",
			)
		}
	} else {
		timeout := webhookTimeoutDurationFunc(s.config.WebhookTimeoutMinutes)
		timeoutChan := time.After(timeout)

		select {
		case <-pipelineDone:
			log.InfoContext(ctx, "Pipeline execution completed, waiting for report",
				"log_kind", "loki_operational",
				"request_id", requestID,
				"date", date,
				"event", "pipeline_done_waiting_report",
			)
			select {
			case r := <-reportReady:
				sendReport(r, true)
			case <-time.After(s.config.EffectiveWebhookReportResultWaitTimeout()):
				log.WarnContext(ctx, "Timeout waiting for report",
					"log_kind", "loki_operational",
					"request_id", requestID,
					"date", date,
					"event", "report_timeout",
				)
			}
		case <-timeoutChan:
			log.WarnContext(ctx, "Webhook timeout reached, sending current status",
				"log_kind", "loki_operational",
				"request_id", requestID,
				"date", date,
				"timeout_minutes", s.config.WebhookTimeoutMinutes,
				"event", "webhook_timeout",
			)
			reportMutex.Lock()
			timeoutReport := &WebhookReport{
				RequestID:          requestID,
				Date:               date,
				StartTime:          startTime,
				Status:             report.Status,
				Success:            report.Success,
				FilesProcessed:     report.FilesProcessed,
				FilesSkipped:       report.FilesSkipped,
				FilesRecovered:     report.FilesRecovered,
				TransactionsLoaded: report.TransactionsLoaded,
				Errors:             report.Errors,
				ErrorMessage:       report.ErrorMessage,
				ErrorBreakdown:     report.ErrorBreakdown,
				ErrorSamples:       report.ErrorSamples,
				KassaDetails:       report.KassaDetails,
			}
			if timeoutReport.Status == "processing" {
				timeoutReport.Status = "timeout"
				timeoutReport.Success = false
				timeoutReport.ErrorMessage = "Pipeline execution timeout reached"
			}
			reportMutex.Unlock()
			sendReport(timeoutReport, false)

			log.InfoContext(ctx, "Timeout report sent, waiting for actual pipeline completion before releasing queue",
				"log_kind", "loki_operational",
				"request_id", requestID,
				"date", date,
				"event", "waiting_pipeline_after_timeout_report",
			)
			<-pipelineDone
			select {
			case r := <-reportReady:
				sendReport(r, true)
			case <-time.After(s.config.EffectiveWebhookReportResultWaitTimeout()):
				log.WarnContext(ctx, "Timeout waiting for final report after pipeline completion",
					"log_kind", "loki_operational",
					"request_id", requestID,
					"date", date,
					"event", "report_timeout_after_pipeline_completion",
				)
			}
		}
	}

	reportMutex.Lock()
	status := report.Status
	duration := time.Since(startTime).String()
	reportMutex.Unlock()

	log.InfoContext(ctx, "ETL pipeline execution finished",
		"log_kind", "loki_operational",
		"request_id", requestID,
		"date", date,
		"status", status,
		"duration", duration,
		"event", "etl_pipeline_execution_finished",
	)

	log.InfoContext(ctx, "Pipeline fully completed, ready for next request",
		"log_kind", "loki_operational",
		"request_id", requestID,
		"date", date,
		"event", "pipeline_fully_completed",
	)
}

// sendWebhookReport отправляет отчет на указанный webhook URL.
func (s *Server) sendWebhookReport(report *WebhookReport) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		s.logger.Error("Error marshaling report",
			"error", err.Error(),
			"event", "report_marshal_error",
		)
		return
	}

	req, err := http.NewRequest("POST", s.config.WebhookReportURL, bytes.NewBuffer(reportJSON))
	if err != nil {
		s.logger.Error("Error creating webhook request",
			"error", err.Error(),
			"event", "webhook_request_error",
		)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Frontol-ETL-Webhook/1.0")
	if s.config.WebhookBearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.WebhookBearerToken)
	}

	client := &http.Client{Timeout: s.config.EffectiveWebhookReportHTTPTimeout()}

	// #nosec G704 -- webhook destination is operator-configured via environment.
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Error sending webhook report",
			"error", err.Error(),
			"event", "webhook_send_error",
		)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Warn("Failed to close webhook response body",
				"error", err.Error(),
				"event", "webhook_response_close_error",
			)
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.logger.Info("Webhook report sent successfully",
			"status_code", resp.StatusCode,
			"event", "webhook_report_sent",
		)
	} else {
		s.logger.Warn("Webhook report failed",
			"status_code", resp.StatusCode,
			"event", "webhook_report_failed",
		)
	}
}
