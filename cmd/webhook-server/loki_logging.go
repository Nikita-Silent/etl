package main

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/user/go-frontol-loader/pkg/logger"
)

type requestAudit struct {
	RequestID string
	Endpoint  string
	Method    string
	Operation string
	ClientIP  string
	UserAgent string
	StartedAt time.Time
}

func newRequestAudit(requestID string, endpoint string, operation string, r *http.Request) requestAudit {
	return requestAudit{
		RequestID: requestID,
		Endpoint:  endpoint,
		Method:    r.Method,
		Operation: operation,
		ClientIP:  requestClientIP(r),
		UserAgent: truncateForLog(r.UserAgent(), 120),
		StartedAt: time.Now(),
	}
}

func (a requestAudit) baseFields() []any {
	return []any{
		"log_kind", "loki_audit",
		"request_id", a.RequestID,
		"endpoint", a.Endpoint,
		"method", a.Method,
		"operation", a.Operation,
		"client_ip", a.ClientIP,
		"user_agent", a.UserAgent,
	}
}

func logAPIRequestReceived(ctx context.Context, log *logger.Logger, audit requestAudit, extra ...any) {
	fields := append(audit.baseFields(), extra...)
	fields = append(fields, "event", "api_request_received")
	log.InfoContext(ctx, "API request received", fields...)
}

func logAPIRequestCompleted(ctx context.Context, log *logger.Logger, audit requestAudit, statusCode int, outcome string, extra ...any) {
	fields := append(audit.baseFields(),
		"status_code", statusCode,
		"outcome", outcome,
		"duration_ms", time.Since(audit.StartedAt).Milliseconds(),
	)
	fields = append(fields, extra...)
	fields = append(fields, "event", "api_request_completed")
	log.InfoContext(ctx, "API request completed", fields...)
}

func logAPIRequestRejected(ctx context.Context, log *logger.Logger, audit requestAudit, statusCode int, reason string, extra ...any) {
	fields := append(audit.baseFields(),
		"status_code", statusCode,
		"outcome", "rejected",
		"reason", reason,
		"duration_ms", time.Since(audit.StartedAt).Milliseconds(),
	)
	fields = append(fields, extra...)
	fields = append(fields, "event", "api_request_rejected")
	log.WarnContext(ctx, "API request rejected", fields...)
}

func requestClientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func truncateForLog(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}
