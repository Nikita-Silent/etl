package logger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Level represents log level
type Level = zerolog.Level

const (
	LevelDebug = zerolog.DebugLevel
	LevelInfo  = zerolog.InfoLevel
	LevelWarn  = zerolog.WarnLevel
	LevelError = zerolog.ErrorLevel
)

// Logger wraps zerolog with a slog-compatible facade for gradual migration.
type Logger struct {
	*slog.Logger
	zlogger zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text/console
	Output io.Writer
	// Filtering (allow only selected values of a structured field).
	FilterField string   // e.g. event or component
	FilterAllow []string // e.g. queue_item_processing_start,queue_item_processing_completed
	TimeFormat  string   // override zerolog time format, defaults to RFC3339Nano
}

// New creates a new Logger with the given configuration
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	applyFilterEnv(&cfg)
	timeFormat := cfg.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339Nano
	}
	zerolog.TimeFieldFormat = timeFormat

	var writer io.Writer = cfg.Output
	allowlist := normalizeAllowlist(cfg.FilterAllow)
	if cfg.FilterField != "" && len(allowlist) > 0 {
		writer = &filteringWriter{
			field: cfg.FilterField,
			allow: allowlist,
			out:   cfg.Output,
		}
	}

	level := parseLevel(cfg.Level)
	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Logger()

	if isTextFormat(cfg.Format) {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: timeFormat,
		}).
			Level(level).
			With().
			Timestamp().
			Logger()
	}

	return newLogger(logger)
}

// Default creates a logger with default settings
func Default() *Logger {
	return New(Config{
		Level:  "info",
		Format: "text",
		Output: os.Stdout,
	})
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// WithRequestID returns a logger with request ID context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return newLogger(l.zlogger.With().Str("request_id", requestID).Logger())
}

// WithComponent returns a logger with component context
func (l *Logger) WithComponent(component string) *Logger {
	return newLogger(l.zlogger.With().Str("component", component).Logger())
}

// WithKassa returns a logger with kassa context
func (l *Logger) WithKassa(kassaCode, folderName string) *Logger {
	return newLogger(l.zlogger.With().
		Str("kassa_code", kassaCode).
		Str("folder", folderName).
		Logger())
}

// ETL logging helpers

// LogETLStart logs ETL pipeline start
func (l *Logger) LogETLStart(ctx context.Context, date string) {
	l.InfoContext(ctx, "ETL pipeline started",
		"date", date,
		"event", "etl_start",
	)
}

// LogETLEnd logs ETL pipeline completion
func (l *Logger) LogETLEnd(ctx context.Context, date string, filesProcessed, transactionsLoaded int, err error) {
	if err != nil {
		l.ErrorContext(ctx, "ETL pipeline failed",
			"date", date,
			"event", "etl_end",
			"files_processed", filesProcessed,
			"transactions_loaded", transactionsLoaded,
			"error", err.Error(),
		)
	} else {
		l.InfoContext(ctx, "ETL pipeline completed",
			"date", date,
			"event", "etl_end",
			"files_processed", filesProcessed,
			"transactions_loaded", transactionsLoaded,
		)
	}
}

// LogFileProcessed logs file processing
func (l *Logger) LogFileProcessed(ctx context.Context, filePath string, transactions int, err error) {
	if err != nil {
		l.ErrorContext(ctx, "File processing failed",
			"file", filePath,
			"event", "file_processed",
			"error", err.Error(),
		)
	} else {
		l.DebugContext(ctx, "File processed",
			"file", filePath,
			"event", "file_processed",
			"transactions", transactions,
		)
	}
}

// LogDBOperation logs database operations
func (l *Logger) LogDBOperation(ctx context.Context, operation, table string, rowsAffected int, err error) {
	if err != nil {
		l.ErrorContext(ctx, "Database operation failed",
			"operation", operation,
			"table", table,
			"event", "db_operation",
			"error", err.Error(),
		)
	} else {
		l.DebugContext(ctx, "Database operation completed",
			"operation", operation,
			"table", table,
			"event", "db_operation",
			"rows_affected", rowsAffected,
		)
	}
}

// LogFTPOperation logs FTP operations
func (l *Logger) LogFTPOperation(ctx context.Context, operation, path string, err error) {
	if err != nil {
		l.ErrorContext(ctx, "FTP operation failed",
			"operation", operation,
			"path", path,
			"event", "ftp_operation",
			"error", err.Error(),
		)
	} else {
		l.DebugContext(ctx, "FTP operation completed",
			"operation", operation,
			"path", path,
			"event", "ftp_operation",
		)
	}
}

// Zerolog exposes the underlying zerolog.Logger.
func (l *Logger) Zerolog() zerolog.Logger {
	return l.zlogger
}

func newLogger(zlogger zerolog.Logger) *Logger {
	handler := &zerologHandler{
		logger: zlogger,
	}

	return &Logger{
		Logger:  slog.New(handler),
		zlogger: zlogger,
	}
}

type zerologHandler struct {
	logger zerolog.Logger
	group  []string
}

func (h *zerologHandler) Enabled(_ context.Context, level slog.Level) bool {
	return levelToZerolog(level) >= h.logger.GetLevel()
}

func (h *zerologHandler) Handle(ctx context.Context, record slog.Record) error {
	event := h.logger.WithLevel(levelToZerolog(record.Level))
	if event == nil {
		return nil
	}

	if ctx != nil {
		event = event.Ctx(ctx)
	}

	record.Attrs(func(attr slog.Attr) bool {
		appendAttr(event, attr, h.group)
		return true
	})

	event.Msg(record.Message)
	return nil
}

func (h *zerologHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	builder := h.logger.With()
	for _, attr := range attrs {
		builder = applyAttr(builder, attr, h.group)
	}

	return &zerologHandler{
		logger: builder.Logger(),
		group:  h.group,
	}
}

func (h *zerologHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroup := append([]string{}, h.group...)
	newGroup = append(newGroup, name)

	return &zerologHandler{
		logger: h.logger,
		group:  newGroup,
	}
}

func appendAttr(event *zerolog.Event, attr slog.Attr, group []string) {
	attr.Value = attr.Value.Resolve()

	switch attr.Value.Kind() {
	case slog.KindGroup:
		groupAttrs := attr.Value.Group()
		for _, nested := range groupAttrs {
			appendAttr(event, nested, append(group, attr.Key))
		}
	default:
		key := joinKey(group, attr.Key)
		applyValue(event, key, attr.Value)
	}
}

func applyAttr(ctx zerolog.Context, attr slog.Attr, group []string) zerolog.Context {
	attr.Value = attr.Value.Resolve()

	switch attr.Value.Kind() {
	case slog.KindGroup:
		for _, nested := range attr.Value.Group() {
			ctx = applyAttr(ctx, nested, append(group, attr.Key))
		}
	default:
		ctx = applyValueToContext(ctx, joinKey(group, attr.Key), attr.Value)
	}

	return ctx
}

func applyValue(event *zerolog.Event, key string, value slog.Value) {
	switch value.Kind() {
	case slog.KindString:
		event.Str(key, value.String())
	case slog.KindInt64:
		event.Int64(key, value.Int64())
	case slog.KindUint64:
		event.Uint64(key, value.Uint64())
	case slog.KindFloat64:
		event.Float64(key, value.Float64())
	case slog.KindBool:
		event.Bool(key, value.Bool())
	case slog.KindTime:
		event.Time(key, value.Time())
	case slog.KindDuration:
		event.Dur(key, value.Duration())
	case slog.KindAny:
		event.Interface(key, value.Any())
	default:
		event.Interface(key, value)
	}
}

func applyValueToContext(ctx zerolog.Context, key string, value slog.Value) zerolog.Context {
	switch value.Kind() {
	case slog.KindString:
		return ctx.Str(key, value.String())
	case slog.KindInt64:
		return ctx.Int64(key, value.Int64())
	case slog.KindUint64:
		return ctx.Uint64(key, value.Uint64())
	case slog.KindFloat64:
		return ctx.Float64(key, value.Float64())
	case slog.KindBool:
		return ctx.Bool(key, value.Bool())
	case slog.KindTime:
		return ctx.Time(key, value.Time())
	case slog.KindDuration:
		return ctx.Dur(key, value.Duration())
	case slog.KindAny:
		return ctx.Interface(key, value.Any())
	default:
		return ctx.Interface(key, value)
	}
}

func levelToZerolog(level slog.Level) zerolog.Level {
	switch {
	case level <= slog.LevelDebug:
		return zerolog.DebugLevel
	case level < slog.LevelWarn:
		return zerolog.InfoLevel
	case level < slog.LevelError:
		return zerolog.WarnLevel
	default:
		return zerolog.ErrorLevel
	}
}

func joinKey(groups []string, key string) string {
	if len(groups) == 0 {
		return key
	}
	return strings.Join(append(groups, key), ".")
}

type filteringWriter struct {
	field string
	allow map[string]struct{}
	out   io.Writer
}

func (fw *filteringWriter) Write(p []byte) (int, error) {
	if fw.field == "" || len(fw.allow) == 0 {
		return fw.out.Write(p)
	}

	var payload map[string]any
	if err := json.Unmarshal(p, &payload); err != nil {
		// If we cannot parse JSON, pass through to avoid losing logs.
		return fw.out.Write(p)
	}

	value, ok := payload[fw.field]
	if !ok {
		return len(p), nil
	}

	if str, ok := value.(string); ok {
		if _, allowed := fw.allow[strings.TrimSpace(str)]; allowed {
			return fw.out.Write(p)
		}
		return len(p), nil
	}

	// Non-string field values bypass filtering.
	return fw.out.Write(p)
}

func normalizeAllowlist(values []string) map[string]struct{} {
	allow := make(map[string]struct{})
	for _, value := range values {
		v := strings.TrimSpace(value)
		if v == "" {
			continue
		}
		allow[v] = struct{}{}
	}
	return allow
}

func isTextFormat(format string) bool {
	switch strings.ToLower(format) {
	case "text", "console":
		return true
	default:
		return false
	}
}

func applyFilterEnv(cfg *Config) {
	if cfg.FilterField == "" {
		cfg.FilterField = strings.TrimSpace(os.Getenv("LOG_FILTER_FIELD"))
	}

	if len(cfg.FilterAllow) == 0 {
		raw := strings.TrimSpace(os.Getenv("LOG_FILTER_ALLOW"))
		if raw != "" {
			cfg.FilterAllow = strings.Split(raw, ",")
		}
	}
}
