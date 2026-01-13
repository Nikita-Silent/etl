// Package server provides HTTP server with graceful shutdown support.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/go-frontol-loader/pkg/logger"
)

// Server represents HTTP server with graceful shutdown
type Server struct {
	httpServer      *http.Server
	logger          *logger.Logger
	shutdownTimeout time.Duration
}

// Config holds server configuration
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig returns default server configuration
func DefaultConfig() Config {
	return Config{
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// New creates a new Server instance
func New(cfg Config, handler http.Handler, log *logger.Logger) *Server {
	if log == nil {
		log = logger.Default()
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		logger:          log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Run starts the server and handles graceful shutdown
func (s *Server) Run(ctx context.Context) error {
	// Channel to receive shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Channel to receive server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("Starting HTTP server",
			"addr", s.httpServer.Addr,
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		s.logger.Info("Shutdown signal received",
			"signal", sig.String(),
		)
	case <-ctx.Done():
		s.logger.Info("Context canceled, initiating shutdown")
	}

	// Graceful shutdown
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.logger.Info("Initiating graceful shutdown",
		"timeout", s.shutdownTimeout.String(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Graceful shutdown failed",
			"error", err.Error(),
		)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}

// HealthCheck represents health check response
type HealthCheck struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Service   string            `json:"service"`
	Version   string            `json:"version,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// HealthChecker defines interface for health checks
type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

// DBHealthChecker checks database health
type DBHealthChecker struct {
	pingFunc func(ctx context.Context) error
}

// NewDBHealthChecker creates a new database health checker
func NewDBHealthChecker(pingFunc func(ctx context.Context) error) *DBHealthChecker {
	return &DBHealthChecker{pingFunc: pingFunc}
}

// Check performs health check
func (c *DBHealthChecker) Check(ctx context.Context) error {
	return c.pingFunc(ctx)
}

// Name returns checker name
func (c *DBHealthChecker) Name() string {
	return "database"
}

// Middleware

// RequestIDMiddleware adds request ID to requests
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			log.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
				"request_id", w.Header().Get("X-Request-ID"),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("Panic recovered",
						"error", fmt.Sprintf("%v", err),
						"path", r.URL.Path,
						"method", r.Method,
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
