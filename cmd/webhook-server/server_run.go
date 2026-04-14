package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/go-frontol-loader/pkg/auth"
	"github.com/user/go-frontol-loader/pkg/server"
)

// Run запускает веб-сервер.
func (s *Server) Run() error {
	bearerAuth := auth.BearerAuthMiddleware(s.logger.Logger, s.config.WebhookBearerToken)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/load", bearerAuth(s.webhookHandler))
	mux.HandleFunc("/api/files", bearerAuth(s.downloadHandler))
	mux.HandleFunc("/api/queue/status", bearerAuth(s.queueStatusHandler))
	mux.HandleFunc("/api/kassas", bearerAuth(s.listKassasHandler))
	mux.HandleFunc("/api/health", s.healthHandler)
	mux.HandleFunc("/api/docs", s.docsHandler)
	mux.HandleFunc("/api/openapi.yaml", s.openAPIHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	handler := server.RequestIDMiddleware(mux)
	handler = server.LoggingMiddleware(s.logger)(handler)
	handler = server.RecoveryMiddleware(s.logger)(handler)

	port := s.config.ServerPort
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf(":%d", port)

	if s.config.WebhookBearerToken == "" {
		s.logger.Warn("Bearer token not configured - authorization is DISABLED",
			"event", "auth_config_missing",
		)
	} else {
		s.logger.Info("Bearer token configured - authorization is ENABLED",
			"token_length", len(s.config.WebhookBearerToken),
			"event", "auth_config_loaded",
		)
	}

	s.logger.Info("Starting webhook server",
		"address", addr,
		"event", "server_start",
	)
	s.logger.Info("Queue system: parallel processing for different operation types, sequential for same operation type",
		"event", "queue_system_info",
	)
	s.logger.Info("Runtime timeout configuration loaded",
		"event", "timeout_config_loaded",
		"db_connect_timeout", s.config.EffectiveDBConnectTimeout(),
		"ftp_connect_timeout", s.config.EffectiveFTPConnectTimeout(),
		"http_read_header_timeout", s.config.EffectiveHTTPReadHeaderTimeout(),
		"http_read_timeout", s.config.EffectiveHTTPReadTimeout(),
		"http_write_timeout", s.config.EffectiveHTTPWriteTimeout(),
		"http_idle_timeout", s.config.EffectiveHTTPIdleTimeout(),
		"pipeline_load_timeout", s.config.EffectivePipelineLoadTimeout(),
		"operation_stale_timeout", s.config.EffectiveOperationStaleTimeout(),
		"webhook_report_http_timeout", s.config.EffectiveWebhookReportHTTPTimeout(),
		"webhook_report_result_wait_timeout", s.config.EffectiveWebhookReportResultWaitTimeout(),
		"shutdown_timeout", s.config.EffectiveShutdownTimeout(),
	)
	if s.opStore != nil {
		abandoned, err := s.opStore.RecoverStale(context.Background())
		if err != nil {
			s.logger.Warn("Failed to recover stale ETL operations",
				"error", err.Error(),
				"event", "operation_recovery_warning",
			)
		} else if abandoned > 0 {
			s.logger.Warn("Recovered stale ETL operations from previous run",
				"abandoned_operations", abandoned,
				"event", "operation_recovery_completed",
			)
		}
	}
	s.logger.Info("Available endpoints",
		"endpoints", []string{
			"POST /api/load - загрузка данных из FTP в БД",
			"GET /api/files?source_folder=XXX&date=YYYY-MM-DD - выгрузка данных из БД в файл",
			"GET /api/queue/status - статус очереди",
			"GET /api/kassas - список касс",
			"GET /api/health - health check",
			"GET /api/docs - документация API (Scalar)",
			"GET /api/openapi.yaml - OpenAPI спецификация",
		},
		"event", "server_endpoints",
	)

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: s.config.EffectiveHTTPReadHeaderTimeout(),
		ReadTimeout:       s.config.EffectiveHTTPReadTimeout(),
		WriteTimeout:      s.config.EffectiveHTTPWriteTimeout(),
		IdleTimeout:       s.config.EffectiveHTTPIdleTimeout(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received",
			"event", "shutdown_signal_received",
		)
		s.Stop()
		return nil
	case err, ok := <-errCh:
		if !ok {
			return nil
		}
		return err
	}
}
