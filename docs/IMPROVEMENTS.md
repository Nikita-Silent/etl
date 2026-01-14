# Comprehensive Improvement Plan for Frontol 6 ETL Loader

## Table of Contents

- [1. Architecture & Extensibility](#1-architecture--extensibility)
- [2. Code Quality & Refactoring](#2-code-quality--refactoring)
- [3. Security Improvements](#3-security-improvements)
- [4. Performance Optimization](#4-performance-optimization)
- [5. Testing Infrastructure](#5-testing-infrastructure)
- [6. Configuration & Deployment](#6-configuration--deployment)
- [7. Monitoring & Observability](#7-monitoring--observability)
- [8. Error Handling & Resilience](#8-error-handling--resilience)
- [9. Developer Experience](#9-developer-experience)
- [10. New Features & Capabilities](#10-new-features--capabilities)

---

## 1. Architecture & Extensibility

### 1.1 Plugin Architecture for Transaction Types

**Problem**: Adding new transaction types requires modifying parser dispatcher and repository code.

**Solution**: Implement plugin-based transaction processor registration.

```go
// pkg/processor/registry.go
type TransactionProcessor interface {
    Type() string
    Parse(fields []string) (interface{}, error)
    TableName() string
    Load(ctx context.Context, tx pgx.Tx, data interface{}) error
    Validate(data interface{}) error
}

type ProcessorRegistry struct {
    processors map[string]TransactionProcessor
}

func (r *ProcessorRegistry) Register(p TransactionProcessor) {
    r.processors[p.Type()] = p
}

// Usage in main.go
registry := processor.NewRegistry()
registry.Register(&processors.TransactionRegistration{})
registry.Register(&processors.BonusTransaction{})
// Auto-discovery via init() in each processor file
```

**Benefits**:
- Add new transaction types without touching core code
- Third-party transaction processors via plugin loading
- Easier testing of individual processors

**Priority**: HIGH - Enables easy feature additions

---

### 1.2 Event-Driven Architecture

**Problem**: Pipeline is monolithic; hard to add pre/post-processing hooks.

**Solution**: Introduce event bus for pipeline stages.

```go
// pkg/events/bus.go
type Event interface {
    Type() string
    Timestamp() time.Time
}

type EventHandler func(ctx context.Context, event Event) error

type EventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

// Events
type FileDownloadedEvent struct { ... }
type FileParsedEvent struct { ... }
type FileLoadedEvent struct { ... }
type PipelineStartedEvent struct { ... }
type PipelineCompletedEvent struct { ... }

// Usage
bus.Subscribe("file.downloaded", notifySlackHandler)
bus.Subscribe("file.loaded", updateMetricsHandler)
bus.Publish(ctx, FileDownloadedEvent{...})
```

**Benefits**:
- Add custom logic without modifying pipeline
- Webhooks, notifications, metrics as event handlers
- Easier to implement retry/compensation logic

**Priority**: MEDIUM

---

### 1.3 Strategy Pattern for File Processing

**Problem**: Different processing strategies hardcoded (parallel vs sequential).

**Solution**: Configurable processing strategies.

```go
// pkg/pipeline/strategy.go
type ProcessingStrategy interface {
    Process(ctx context.Context, tasks []FileTask) error
}

type ParallelStrategy struct {
    MaxConcurrency int
}

type SequentialStrategy struct{}

type BatchedStrategy struct {
    BatchSize int
}

// Configuration
type PipelineConfig struct {
    Strategy ProcessingStrategy
}
```

**Benefits**:
- Easy to add new processing modes (priority-based, rate-limited)
- Per-kassa strategy configuration
- A/B testing different strategies

**Priority**: MEDIUM

---

### 1.4 Repository Pattern Enhancement

**Problem**: Repository tightly coupled to pgx; hard to add caching layer.

**Solution**: Add repository abstraction with decorators.

```go
// pkg/repository/interfaces.go
type TransactionRepository interface {
    LoadTransactions(ctx context.Context, txType string, data []interface{}) error
    GetTransactionsByDate(ctx context.Context, date time.Time) ([]interface{}, error)
}

// pkg/repository/decorators.go
type CachedRepository struct {
    base  TransactionRepository
    cache cache.Cache
}

type MetricsRepository struct {
    base    TransactionRepository
    metrics *prometheus.Registry
}

type RetryRepository struct {
    base       TransactionRepository
    maxRetries int
}
```

**Benefits**:
- Add caching without changing core logic
- Metrics collection via decorator
- Retry logic reusable across repositories

**Priority**: HIGH

---

### 1.5 Service Layer Separation

**Problem**: Business logic mixed with infrastructure (FTP, DB).

**Solution**: Introduce service layer.

```
pkg/
├── domain/           # Business entities
│   ├── transaction.go
│   ├── kassa.go
│   └── file.go
├── services/         # Business logic
│   ├── etl_service.go
│   ├── file_service.go
│   └── validation_service.go
├── infrastructure/   # External dependencies
│   ├── ftp/
│   ├── database/
│   └── storage/
└── application/      # Use cases
    ├── load_data.go
    └── export_data.go
```

**Benefits**:
- Business logic testable without infrastructure
- Clear separation of concerns
- Easier to replace infrastructure components

**Priority**: HIGH

---

### 1.6 Configuration Schema Versioning

**Problem**: KASSA_STRUCTURE format changes break existing configs.

**Solution**: Versioned configuration with migration.

```go
// pkg/config/schema.go
type ConfigSchema interface {
    Version() int
    Migrate(from ConfigSchema) error
}

type ConfigV1 struct {
    KassaStructure string
}

type ConfigV2 struct {
    Kassas []KassaConfig
}

// Auto-migration on load
func LoadConfig() (*Config, error) {
    raw := loadRawConfig()
    schema := detectSchema(raw)
    if schema.Version() < CurrentVersion {
        schema = migrateSchema(schema)
    }
    return parseConfig(schema)
}
```

**Priority**: MEDIUM

---

## 2. Code Quality & Refactoring

### 2.1 Remove Duplicate processFile Functions

**Location**: `pkg/pipeline/pipeline.go:387-584`

**Action**:
```go
// Remove processFile() entirely
// Keep only processFileWithMutex()
// Update all callers
```

**Impact**: -200 lines, easier maintenance

**Priority**: HIGH (Quick Win)

---

### 2.2 Extract FTP Operations to Service

**Problem**: FTP logic scattered across pipeline.

**Solution**:
```go
// pkg/services/ftp_service.go
type FTPService struct {
    client FTPClient
    mu     sync.Mutex
}

func (s *FTPService) ListUnprocessedFiles(ctx context.Context, folder FolderConfig) ([]File, error)
func (s *FTPService) DownloadAndMarkProcessed(ctx context.Context, file File) ([]byte, error)
func (s *FTPService) CleanupProcessedFiles(ctx context.Context, olderThan time.Duration) error
```

**Priority**: MEDIUM

---

### 2.3 Unified Error Types

**Problem**: Error handling inconsistent across packages.

**Solution**:
```go
// pkg/errors/errors.go
type ErrorCode string

const (
    ErrCodeFTPConnection    ErrorCode = "FTP_CONNECTION"
    ErrCodeParseInvalid     ErrorCode = "PARSE_INVALID"
    ErrCodeDBDeadlock       ErrorCode = "DB_DEADLOCK"
    ErrCodeValidation       ErrorCode = "VALIDATION"
)

type AppError struct {
    Code    ErrorCode
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *AppError) Is(target error) bool {
    t, ok := target.(*AppError)
    return ok && t.Code == e.Code
}

// Usage
if errors.Is(err, &AppError{Code: ErrCodeDBDeadlock}) {
    // Handle specifically
}
```

**Benefits**:
- Structured error handling
- Error categorization for metrics
- Better error messages for users

**Priority**: HIGH

---

### 2.4 Context Propagation Fix

**Location**: `pkg/pipeline/pipeline.go:449-452`

**Current**:
```go
loadCtx, loadCancel := context.WithTimeout(context.Background(), 1*time.Hour)
```

**Fix**:
```go
loadCtx, loadCancel := context.WithTimeout(ctx, 1*time.Hour)
```

**Priority**: HIGH (Bug Fix)

---

### 2.5 Interface Segregation

**Problem**: Large interfaces hard to mock.

**Solution**: Split into smaller interfaces.

```go
// Current (large)
type DatabasePool interface {
    BeginTx(...) (pgx.Tx, error)
    Query(...) (pgx.Rows, error)
    LoadData(...) error
    // ... 20 methods
}

// Better (segregated)
type TransactionManager interface {
    BeginTx(ctx context.Context) (pgx.Tx, error)
    Commit(ctx context.Context, tx pgx.Tx) error
    Rollback(ctx context.Context, tx pgx.Tx) error
}

type DataLoader interface {
    LoadTransactions(ctx context.Context, tx pgx.Tx, data []interface{}) error
}

type QueryExecutor interface {
    Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}
```

**Priority**: MEDIUM

---

### 2.6 Builder Pattern for Configuration

**Problem**: Config initialization complex with many optional fields.

**Solution**:
```go
// pkg/config/builder.go
type ConfigBuilder struct {
    config *Config
}

func NewConfigBuilder() *ConfigBuilder {
    return &ConfigBuilder{config: &Config{}}
}

func (b *ConfigBuilder) WithFTP(host string, port int) *ConfigBuilder {
    b.config.FTPHost = host
    b.config.FTPPort = port
    return b
}

func (b *ConfigBuilder) WithDatabase(dsn string) *ConfigBuilder {
    b.config.DBDSN = dsn
    return b
}

func (b *ConfigBuilder) Build() (*Config, error) {
    if err := b.validate(); err != nil {
        return nil, err
    }
    return b.config, nil
}

// Usage
config, err := NewConfigBuilder().
    WithFTP("ftp.example.com", 21).
    WithDatabase("postgres://...").
    WithKassas(kassas).
    Build()
```

**Priority**: LOW

---

### 2.7 Value Objects for Domain Concepts

**Problem**: Primitive obsession (strings for dates, folder names).

**Solution**:
```go
// pkg/domain/values.go
type Date struct {
    time.Time
}

func NewDate(s string) (Date, error) {
    t, err := time.Parse("2006-01-02", s)
    if err != nil {
        return Date{}, fmt.Errorf("invalid date format: %w", err)
    }
    return Date{t}, nil
}

type KassaCode struct {
    value string
}

func NewKassaCode(s string) (KassaCode, error) {
    if !kassaCodeRegex.MatchString(s) {
        return KassaCode{}, errors.New("invalid kassa code")
    }
    return KassaCode{s}, nil
}

type FolderPath struct {
    value string
}
```

**Benefits**:
- Validation at construction
- Type safety (can't pass KassaCode where FolderPath expected)
- Self-documenting code

**Priority**: MEDIUM

---

## 3. Security Improvements

### 3.1 Constant-Time Token Comparison

**Location**: `pkg/auth/auth.go:75`

**Current**:
```go
if providedToken != token {
    return false
}
```

**Fix**:
```go
import "crypto/subtle"

if subtle.ConstantTimeCompare([]byte(providedToken), []byte(token)) != 1 {
    return false
}
```

**Priority**: HIGH (Quick Win)

---

### 3.2 Rate Limiting Middleware

**Problem**: No protection against DOS attacks.

**Solution**:
```go
// pkg/middleware/ratelimit.go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := getClientIP(r)
        limiter := rl.getLimiter(ip)

        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Configuration
RATE_LIMIT_REQUESTS_PER_SECOND=10
RATE_LIMIT_BURST=20
RATE_LIMIT_BY_IP=true
RATE_LIMIT_BY_TOKEN=true
```

**Priority**: HIGH

---

### 3.3 Input Validation Framework

**Problem**: Validation scattered, inconsistent.

**Solution**:
```go
// pkg/validation/validator.go
type Validator interface {
    Validate(v interface{}) error
}

type DateValidator struct{}

func (dv DateValidator) Validate(v interface{}) error {
    date, ok := v.(string)
    if !ok {
        return errors.New("date must be string")
    }
    _, err := time.Parse("2006-01-02", date)
    return err
}

type KassaCodeValidator struct{}

type CompositeValidator struct {
    validators []Validator
}

// Usage in handlers
validator := validation.NewComposite(
    validation.Required(),
    validation.DateFormat("2006-01-02"),
    validation.NotInFuture(),
)

if err := validator.Validate(request.Date); err != nil {
    return BadRequest(err)
}
```

**Priority**: HIGH

---

### 3.4 Secret Management

**Problem**: Passwords in plaintext environment variables.

**Solution**:
```go
// pkg/secrets/manager.go
type SecretManager interface {
    GetSecret(ctx context.Context, key string) (string, error)
}

type VaultSecretManager struct {
    client *vault.Client
}

type AWSSecretsManager struct {
    client *secretsmanager.Client
}

type EnvSecretManager struct{} // Fallback for development

// Configuration
SECRET_BACKEND=vault|aws|env
VAULT_ADDR=https://vault.example.com
AWS_REGION=us-east-1
```

**Priority**: MEDIUM (for production)

---

### 3.5 Audit Logging

**Problem**: No audit trail of operations.

**Solution**:
```go
// pkg/audit/logger.go
type AuditLog struct {
    Timestamp time.Time
    UserID    string
    IP        string
    Operation string
    Resource  string
    Status    string
    Details   map[string]interface{}
}

type AuditLogger interface {
    Log(ctx context.Context, log AuditLog) error
}

// Storage in separate audit_logs table
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    user_id TEXT,
    ip_address INET,
    operation TEXT NOT NULL,
    resource TEXT,
    status TEXT,
    details JSONB,
    INDEX idx_audit_timestamp (timestamp),
    INDEX idx_audit_operation (operation)
);
```

**Priority**: MEDIUM

---

### 3.6 SQL Injection Prevention

**Problem**: Table names in string formatting.

**Current**:
```go
query := fmt.Sprintf("INSERT INTO %s ...", tableName)
```

**Solution**:
```go
// pkg/db/safe_query.go
var allowedTables = map[string]bool{
    "tx_item_registration_1_11": true,
    "tx_bonus_accrual_9": true,
    // ...
}

func SafeTableName(name string) (string, error) {
    if !allowedTables[name] {
        return "", errors.New("invalid table name")
    }
    return name, nil
}

// Or use pgx identifier quoting
import "github.com/jackc/pgx/v5"

query := fmt.Sprintf("INSERT INTO %s ...", pgx.Identifier{tableName}.Sanitize())
```

**Priority**: MEDIUM

---

### 3.7 TLS for FTP Connections

**Problem**: FTP connections unencrypted.

**Solution**:
```go
// pkg/ftp/ftp.go
import "github.com/jlaffaye/ftp"

func NewClient(cfg Config) (*Client, error) {
    var conn *ftp.ServerConn
    var err error

    if cfg.FTPUseTLS {
        tlsConfig := &tls.Config{
            InsecureSkipVerify: cfg.FTPInsecureSkipVerify,
        }
        conn, err = ftp.Dial(
            cfg.FTPHost+":"+cfg.FTPPort,
            ftp.DialWithTLS(tlsConfig),
        )
    } else {
        conn, err = ftp.Dial(cfg.FTPHost + ":" + cfg.FTPPort)
    }

    // ...
}

// Configuration
FTP_USE_TLS=true
FTP_INSECURE_SKIP_VERIFY=false
```

**Priority**: HIGH (for production)

---

### 3.8 Request Signature Verification

**Problem**: Webhook calls can be spoofed.

**Solution**:
```go
// pkg/webhook/signature.go
import "crypto/hmac"
import "crypto/sha256"

func VerifySignature(body []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) == 1
}

// Usage in handler
signature := r.Header.Get("X-Webhook-Signature")
body, _ := io.ReadAll(r.Body)
if !webhook.VerifySignature(body, signature, cfg.WebhookSecret) {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
}
```

**Priority**: MEDIUM

---

## 4. Performance Optimization

### 4.1 FTP Connection Pooling

**Problem**: Single FTP connection with mutex bottleneck.

**Solution**:
```go
// pkg/ftp/pool.go
type ConnectionPool struct {
    conns   chan *ftp.ServerConn
    factory func() (*ftp.ServerConn, error)
    maxSize int
}

func NewConnectionPool(maxSize int, factory func() (*ftp.ServerConn, error)) *ConnectionPool {
    pool := &ConnectionPool{
        conns:   make(chan *ftp.ServerConn, maxSize),
        factory: factory,
        maxSize: maxSize,
    }

    // Pre-warm pool
    for i := 0; i < maxSize; i++ {
        conn, err := factory()
        if err != nil {
            continue
        }
        pool.conns <- conn
    }

    return pool
}

func (p *ConnectionPool) Get(ctx context.Context) (*ftp.ServerConn, error) {
    select {
    case conn := <-p.conns:
        return conn, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        return p.factory()
    }
}

func (p *ConnectionPool) Put(conn *ftp.ServerConn) {
    select {
    case p.conns <- conn:
    default:
        conn.Quit()
    }
}
```

**Impact**: 5-10x throughput improvement for parallel downloads

**Priority**: HIGH

---

### 4.2 Worker Pool for File Processing

**Problem**: Unbounded goroutine creation.

**Current**:
```go
for _, task := range tasks {
    wg.Add(1)
    go func(t fileTask) { ... }(task)
}
```

**Solution**:
```go
// pkg/workers/pool.go
import "golang.org/x/sync/semaphore"

type WorkerPool struct {
    sem *semaphore.Weighted
}

func NewWorkerPool(size int) *WorkerPool {
    return &WorkerPool{
        sem: semaphore.NewWeighted(int64(size)),
    }
}

func (wp *WorkerPool) Submit(ctx context.Context, fn func()) error {
    if err := wp.sem.Acquire(ctx, 1); err != nil {
        return err
    }

    go func() {
        defer wp.sem.Release(1)
        fn()
    }()

    return nil
}

// Usage in pipeline
pool := workers.NewWorkerPool(runtime.NumCPU() * 2)
for _, task := range tasks {
    t := task
    pool.Submit(ctx, func() {
        processFile(t)
    })
}

// Configuration
WORKER_POOL_SIZE=16
```

**Impact**: Controlled memory usage, better CPU utilization

**Priority**: HIGH

---

### 4.3 Batch FTP Listing

**Problem**: Multiple FTP LIST calls for processed files.

**Current**:
```go
for _, file := range files {
    if ftpClient.IsFileProcessed(file) {
        continue
    }
}
```

**Solution**:
```go
// pkg/ftp/batch.go
func (c *Client) ListUnprocessedFiles(folder string) ([]FileInfo, error) {
    // Single LIST call
    allFiles, err := c.conn.List(folder)
    if err != nil {
        return nil, err
    }

    // Filter in-memory
    var unprocessed []FileInfo
    processedMarkers := make(map[string]bool)

    for _, f := range allFiles {
        if strings.HasSuffix(f.Name, ".processed") {
            baseName := strings.TrimSuffix(f.Name, ".processed")
            processedMarkers[baseName] = true
        }
    }

    for _, f := range allFiles {
        if !processedMarkers[f.Name] && !strings.HasSuffix(f.Name, ".processed") {
            unprocessed = append(unprocessed, f)
        }
    }

    return unprocessed, nil
}
```

**Impact**: O(n) instead of O(n²) FTP operations

**Priority**: HIGH

---

### 4.4 Operation Type Caching

**Problem**: DB query for each operation type validation.

**Solution**:
```go
// pkg/repository/operation_cache.go
type OperationTypeCache struct {
    types map[int]bool
    mu    sync.RWMutex
    ttl   time.Duration
}

func NewOperationTypeCache(db DatabasePool) *OperationTypeCache {
    cache := &OperationTypeCache{
        types: make(map[int]bool),
        ttl:   1 * time.Hour,
    }

    // Load on init
    cache.refresh(context.Background(), db)

    // Periodic refresh
    go func() {
        ticker := time.NewTicker(cache.ttl)
        for range ticker.C {
            cache.refresh(context.Background(), db)
        }
    }()

    return cache
}

func (c *OperationTypeCache) Exists(opType int) bool {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.types[opType]
}
```

**Impact**: Eliminates thousands of DB queries per ETL run

**Priority**: HIGH

---

### 4.5 Parallel Database Inserts

**Problem**: Sequential inserts for large files.

**Solution**:
```go
// pkg/repository/parallel_loader.go
func (l *Loader) LoadDataParallel(ctx context.Context, tableName string, data []map[string]interface{}) error {
    batchSize := 1000
    numWorkers := 4

    batches := splitIntoBatches(data, batchSize)
    errChan := make(chan error, len(batches))

    sem := semaphore.NewWeighted(int64(numWorkers))
    for _, batch := range batches {
        b := batch
        sem.Acquire(ctx, 1)

        go func() {
            defer sem.Release(1)

            tx, err := l.db.BeginTx(ctx)
            if err != nil {
                errChan <- err
                return
            }

            err = l.insertBatch(ctx, tx, tableName, b)
            if err != nil {
                tx.Rollback(ctx)
                errChan <- err
                return
            }

            errChan <- tx.Commit(ctx)
        }()
    }

    // Collect errors
    for i := 0; i < len(batches); i++ {
        if err := <-errChan; err != nil {
            return err
        }
    }

    return nil
}
```

**Impact**: 2-4x faster for large files (10K+ transactions)

**Priority**: MEDIUM

---

### 4.6 Database Connection Pool Tuning

**Problem**: Hardcoded pool sizes.

**Solution**:
```go
// pkg/db/postgres.go
func NewPool(cfg *Config) (*Pool, error) {
    config, err := pgxpool.ParseConfig(cfg.DBDSN)
    if err != nil {
        return nil, err
    }

    // Configurable pool settings
    config.MaxConns = int32(cfg.DBMaxConns)
    config.MinConns = int32(cfg.DBMinConns)
    config.MaxConnLifetime = cfg.DBMaxConnLifetime
    config.MaxConnIdleTime = cfg.DBMaxConnIdleTime
    config.HealthCheckPeriod = cfg.DBHealthCheckPeriod

    // Connection timeout
    config.ConnConfig.ConnectTimeout = cfg.DBConnectTimeout

    // Statement cache
    config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement

    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        return nil, err
    }

    return &Pool{pool: pool}, nil
}

// Configuration
DB_MAX_CONNS=20
DB_MIN_CONNS=5
DB_MAX_CONN_LIFETIME=1h
DB_MAX_CONN_IDLE_TIME=30m
DB_HEALTH_CHECK_PERIOD=1m
DB_CONNECT_TIMEOUT=5s
```

**Priority**: MEDIUM

---

### 4.7 Lazy Loading for Large Files

**Problem**: Entire file loaded into memory before processing.

**Solution**:
```go
// pkg/parser/streaming.go
type StreamingParser struct {
    reader io.Reader
}

func (p *StreamingParser) ParseStream(ctx context.Context, handler func(Transaction) error) error {
    scanner := bufio.NewScanner(p.reader)
    scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 1MB max line

    lineNum := 0
    for scanner.Scan() {
        lineNum++

        if ctx.Err() != nil {
            return ctx.Err()
        }

        line := scanner.Text()
        tx, err := p.parseLine(line)
        if err != nil {
            return fmt.Errorf("line %d: %w", lineNum, err)
        }

        if err := handler(tx); err != nil {
            return err
        }
    }

    return scanner.Err()
}

// Usage with batching
batch := make([]Transaction, 0, 1000)
parser.ParseStream(ctx, func(tx Transaction) error {
    batch = append(batch, tx)

    if len(batch) >= 1000 {
        if err := loader.LoadBatch(ctx, batch); err != nil {
            return err
        }
        batch = batch[:0]
    }

    return nil
})
```

**Impact**: Constant memory usage regardless of file size

**Priority**: MEDIUM

---

### 4.8 COPY Instead of INSERT for Bulk Loads

**Problem**: INSERT slower than PostgreSQL COPY for bulk data.

**Solution**:
```go
// pkg/repository/copy_loader.go
func (l *Loader) LoadDataWithCopy(ctx context.Context, tableName string, data []map[string]interface{}) error {
    tx, err := l.db.BeginTx(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    columns := getColumns(data[0])

    // Use COPY protocol
    copyCount, err := tx.CopyFrom(
        ctx,
        pgx.Identifier{tableName},
        columns,
        pgx.CopyFromRows(data),
    )
    if err != nil {
        return err
    }

    log.Info("Copied %d rows", copyCount)

    return tx.Commit(ctx)
}
```

**Impact**: 5-10x faster than individual INSERTs

**Priority**: HIGH

---

### 4.9 Compiled Regular Expressions

**Problem**: Regex compiled on every parse.

**Solution**:
```go
// pkg/parser/parser.go
var (
    kassaCodeRegex = regexp.MustCompile(`^[0-9]{3}$`)
    dateRegex      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
    // Compile once at package init
)

func init() {
    // All regex compilation happens here
}
```

**Priority**: LOW (micro-optimization)

---

### 4.10 Index Optimization

**Problem**: Missing indexes for common queries.

**Solution**:
```sql
-- Add composite indexes for common queries
CREATE INDEX CONCURRENTLY idx_transactions_date_kassa
    ON tx_item_registration_1_11(transaction_date, source_folder);

CREATE INDEX CONCURRENTLY idx_transactions_lookup
    ON tx_item_registration_1_11(transaction_id_unique, source_folder)
    WHERE processed_at > NOW() - INTERVAL '7 days';

-- Partial indexes for recent data
CREATE INDEX CONCURRENTLY idx_recent_transactions
    ON tx_item_registration_1_11(transaction_date)
    WHERE transaction_date > CURRENT_DATE - INTERVAL '30 days';

-- Expression indexes for common filters
CREATE INDEX CONCURRENTLY idx_transaction_status
    ON tx_item_registration_1_11((status::text))
    WHERE status IS NOT NULL;
```

**Priority**: MEDIUM

---

## 5. Testing Infrastructure

### 5.1 Integration Test Framework

**Problem**: Integration tests hard to write and maintain.

**Solution**:
```go
// tests/integration/framework/framework.go
type TestEnvironment struct {
    DB      *pgxpool.Pool
    FTP     *ftp.Client
    Config  *config.Config
    Cleanup []func()
}

func NewTestEnvironment(t *testing.T) *TestEnvironment {
    t.Helper()

    // Start test containers
    dbContainer := startPostgresContainer(t)
    ftpContainer := startFTPContainer(t)

    env := &TestEnvironment{
        DB:  connectToDB(dbContainer),
        FTP: connectToFTP(ftpContainer),
    }

    env.Cleanup = append(env.Cleanup, func() {
        dbContainer.Terminate(context.Background())
        ftpContainer.Terminate(context.Background())
    })

    t.Cleanup(func() {
        for _, fn := range env.Cleanup {
            fn()
        }
    })

    return env
}

func (e *TestEnvironment) SeedDatabase(fixtures ...Fixture) error {
    for _, f := range fixtures {
        if err := f.Load(e.DB); err != nil {
            return err
        }
    }
    return nil
}

func (e *TestEnvironment) SeedFTP(files ...File) error {
    for _, f := range files {
        if err := e.FTP.Upload(f); err != nil {
            return err
        }
    }
    return nil
}

// Usage
func TestFullETLPipeline(t *testing.T) {
    env := framework.NewTestEnvironment(t)

    env.SeedFTP(
        fixtures.FrontolFile("kassa001", "2024-12-18.txt", 100),
    )

    pipeline := pipeline.New(env.Config, env.DB, env.FTP)
    err := pipeline.Run(context.Background(), "2024-12-18")

    assert.NoError(t, err)
    assertTransactionsLoaded(t, env.DB, 100)
}
```

**Priority**: HIGH

---

### 5.2 Test Data Builders

**Problem**: Test data creation verbose and error-prone.

**Solution**:
```go
// tests/builders/transaction.go
type TransactionBuilder struct {
    tx models.TransactionRegistration
}

func NewTransaction() *TransactionBuilder {
    return &TransactionBuilder{
        tx: models.TransactionRegistration{
            TransactionIDUnique: uuid.NewString(),
            SourceFolder:        "001",
            TransactionDate:     time.Now(),
            // ... defaults
        },
    }
}

func (b *TransactionBuilder) WithKassa(code string) *TransactionBuilder {
    b.tx.SourceFolder = code
    return b
}

func (b *TransactionBuilder) WithDate(date string) *TransactionBuilder {
    b.tx.TransactionDate, _ = time.Parse("2006-01-02", date)
    return b
}

func (b *TransactionBuilder) Build() models.TransactionRegistration {
    return b.tx
}

// Usage
tx := builders.NewTransaction().
    WithKassa("001").
    WithDate("2024-12-18").
    WithAmount(1500).
    Build()
```

**Priority**: MEDIUM

---

### 5.3 Table-Driven Tests

**Problem**: Repetitive test cases.

**Solution**:
```go
// pkg/parser/parser_test.go
func TestParser_ParseLine(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    models.Transaction
        wantErr bool
    }{
        {
            name:  "valid transaction type 1",
            input: "001;123;2024-12-18;1;...",
            want:  models.TransactionRegistration{...},
        },
        {
            name:    "invalid date format",
            input:   "001;123;18-12-2024;1;...",
            wantErr: true,
        },
        {
            name:    "missing fields",
            input:   "001;123",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parser.ParseLine(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Priority**: MEDIUM

---

### 5.4 Benchmark Suite

**Problem**: No performance regression detection.

**Solution**:
```go
// pkg/parser/parser_bench_test.go
func BenchmarkParser_ParseFile(b *testing.B) {
    data := generateTestFile(10000) // 10K transactions

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parser.ParseFile(bytes.NewReader(data))
    }
}

func BenchmarkRepository_LoadData(b *testing.B) {
    db := setupTestDB(b)
    data := generateTransactions(1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        repo.LoadData(context.Background(), data)
    }
}

// Run benchmarks with memory profiling
// go test -bench=. -benchmem -cpuprofile=cpu.prof
```

**Priority**: MEDIUM

---

### 5.5 Mutation Testing

**Problem**: Tests may not catch all bugs.

**Solution**:
```bash
# Use go-mutesting
go install github.com/zimmski/go-mutesting/cmd/go-mutesting@latest

# Run mutation tests
go-mutesting ./pkg/...

# Report shows which mutations survived (weak tests)
```

**Priority**: LOW

---

### 5.6 Contract Testing

**Problem**: FTP/DB interfaces may drift from implementations.

**Solution**:
```go
// pkg/ftp/contract_test.go
func TestFTPClientContract(t *testing.T) {
    implementations := []struct {
        name   string
        client FTPClient
    }{
        {"Real", NewClient(realConfig)},
        {"Mock", NewMockClient()},
    }

    for _, impl := range implementations {
        t.Run(impl.name, func(t *testing.T) {
            runContractTests(t, impl.client)
        })
    }
}

func runContractTests(t *testing.T, client FTPClient) {
    t.Run("ListFiles returns sorted files", func(t *testing.T) {
        files, err := client.ListFiles("/")
        assert.NoError(t, err)
        assert.True(t, sort.IsSorted(byModTime(files)))
    })

    t.Run("DownloadFile returns error for missing file", func(t *testing.T) {
        _, err := client.DownloadFile("nonexistent")
        assert.Error(t, err)
    })
}
```

**Priority**: MEDIUM

---

### 5.7 E2E Tests with Testcontainers

**Problem**: E2E tests require manual setup.

**Solution**:
```go
// tests/e2e/e2e_test.go
import "github.com/testcontainers/testcontainers-go"

func TestE2E_FullPipeline(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    ctx := context.Background()

    // Start PostgreSQL
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:16",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)

    // Start FTP server
    ftpServer := startFTPTestContainer(t, ctx)
    defer ftpServer.Terminate(ctx)

    // Build and start webhook server
    webhookServer := startWebhookServer(t, postgres, ftpServer)
    defer webhookServer.Stop()

    // Seed FTP with test files
    seedFTP(t, ftpServer, "testdata/frontol_export.txt")

    // Trigger ETL via webhook
    resp := triggerETL(t, webhookServer.URL, "2024-12-18")
    assert.Equal(t, http.StatusAccepted, resp.StatusCode)

    // Wait for completion
    waitForCompletion(t, webhookServer.URL, resp.RequestID, 30*time.Second)

    // Verify database state
    assertDatabaseContains(t, postgres, 100, "tx_item_registration_1_11")
}
```

**Priority**: HIGH

---

### 5.8 Fuzzing for Parsers

**Problem**: Edge cases in parsing not covered.

**Solution**:
```go
// pkg/parser/parser_fuzz_test.go
func FuzzParseLine(f *testing.F) {
    // Seed corpus
    f.Add("001;123;2024-12-18;1;field1;field2")
    f.Add("002;456;2024-12-19;2;")

    f.Fuzz(func(t *testing.T, input string) {
        // Should never panic
        _, err := parser.ParseLine(input)

        // If valid, should round-trip
        if err == nil {
            // Verify data integrity
        }
    })
}

// Run fuzzing
// go test -fuzz=FuzzParseLine -fuzztime=30s
```

**Priority**: MEDIUM

---

### 5.9 Property-Based Testing

**Problem**: Hard to test all edge cases manually.

**Solution**:
```go
// pkg/repository/loader_property_test.go
import "github.com/leanovate/gopter"

func TestLoader_IdempotencyProperty(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("Loading same data twice produces same result", prop.ForAll(
        func(transactions []models.Transaction) bool {
            db := setupTestDB(t)

            // Load first time
            err1 := loader.LoadData(ctx, transactions)
            count1 := getRowCount(db)

            // Load second time (should be idempotent)
            err2 := loader.LoadData(ctx, transactions)
            count2 := getRowCount(db)

            return err1 == nil && err2 == nil && count1 == count2
        },
        gen.SliceOf(genTransaction()),
    ))

    properties.TestingRun(t)
}
```

**Priority**: LOW

---

## 6. Configuration & Deployment

### 6.1 Multi-Environment Configuration

**Problem**: Single config for all environments.

**Solution**:
```go
// pkg/config/environment.go
type Environment string

const (
    EnvDevelopment Environment = "development"
    EnvStaging     Environment = "staging"
    EnvProduction  Environment = "production"
)

type EnvironmentConfig struct {
    Base       Config
    Overrides  map[Environment]Config
}

func LoadConfig() (*Config, error) {
    env := Environment(os.Getenv("APP_ENV"))
    if env == "" {
        env = EnvDevelopment
    }

    cfg := loadBaseConfig()

    // Apply environment-specific overrides
    if override, ok := cfg.Overrides[env]; ok {
        cfg = mergeConfigs(cfg.Base, override)
    }

    return &cfg, nil
}

// File structure
configs/
├── base.env
├── development.env
├── staging.env
└── production.env
```

**Priority**: MEDIUM

---

### 6.2 Configuration Validation on Startup

**Problem**: Invalid config discovered at runtime.

**Solution**:
```go
// pkg/config/validator.go
func (c *Config) Validate() error {
    var errs []error

    if c.FTPHost == "" {
        errs = append(errs, errors.New("FTP_HOST is required"))
    }

    if c.FTPPort < 1 || c.FTPPort > 65535 {
        errs = append(errs, errors.New("FTP_PORT must be 1-65535"))
    }

    if len(c.KassaStructure) == 0 {
        errs = append(errs, errors.New("KASSA_STRUCTURE cannot be empty"))
    }

    if c.DBMaxConns < c.DBMinConns {
        errs = append(errs, errors.New("DB_MAX_CONNS must be >= DB_MIN_CONNS"))
    }

    if len(errs) > 0 {
        return fmt.Errorf("configuration validation failed:\n%s", joinErrors(errs))
    }

    return nil
}

// In main.go
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal(err)
}

if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

**Priority**: HIGH

---

### 6.3 Feature Flags

**Problem**: Hard to enable/disable features without code changes.

**Solution**:
```go
// pkg/features/flags.go
type FeatureFlags struct {
    EnableParallelProcessing bool
    EnableMetricsExport      bool
    EnableAuditLogging       bool
    EnableRateLimiting       bool
    EnableCaching            bool
    MaxFileSize              int64
    BatchInsertSize          int
}

func LoadFeatureFlags() *FeatureFlags {
    return &FeatureFlags{
        EnableParallelProcessing: getEnvAsBool("FEATURE_PARALLEL_PROCESSING", true),
        EnableMetricsExport:      getEnvAsBool("FEATURE_METRICS_EXPORT", false),
        EnableAuditLogging:       getEnvAsBool("FEATURE_AUDIT_LOGGING", true),
        MaxFileSize:              getEnvAsInt64("FEATURE_MAX_FILE_SIZE", 100*1024*1024),
        BatchInsertSize:          getEnvAsInt("FEATURE_BATCH_INSERT_SIZE", 1000),
    }
}

// Usage in code
if cfg.Features.EnableParallelProcessing {
    processInParallel(tasks)
} else {
    processSequentially(tasks)
}
```

**Priority**: MEDIUM

---

### 6.4 Graceful Shutdown Improvements

**Problem**: 30s timeout may be too short.

**Solution**:
```go
// pkg/server/shutdown.go
func (s *Server) GracefulShutdown(ctx context.Context) error {
    shutdownTimeout := s.cfg.ShutdownTimeout
    if shutdownTimeout == 0 {
        shutdownTimeout = 30 * time.Second
    }

    shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
    defer cancel()

    // Stop accepting new requests
    s.logger.Info("Stopping HTTP server...")
    if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
        return fmt.Errorf("HTTP shutdown: %w", err)
    }

    // Drain queue
    s.logger.Info("Draining request queue...")
    s.requestManager.StopAcceptingRequests()

    // Wait for workers with timeout
    done := make(chan struct{})
    go func() {
        s.workerWg.Wait()
        close(done)
    }()

    select {
    case <-done:
        s.logger.Info("All workers finished gracefully")
    case <-shutdownCtx.Done():
        s.logger.Warn("Shutdown timeout exceeded, forcing termination")
        return errors.New("shutdown timeout")
    }

    // Close database connections
    s.logger.Info("Closing database connections...")
    s.db.Close()

    return nil
}

// Configuration
SHUTDOWN_TIMEOUT_SECONDS=60
QUEUE_DRAIN_TIMEOUT_SECONDS=120
```

**Priority**: HIGH

---

### 6.5 Health Check Enhancements

**Problem**: Health check doesn't verify dependencies.

**Solution**:
```go
// pkg/health/checker.go
type HealthCheck struct {
    Status      string                 `json:"status"`
    Version     string                 `json:"version"`
    Uptime      string                 `json:"uptime"`
    Checks      map[string]CheckResult `json:"checks"`
    Timestamp   time.Time              `json:"timestamp"`
}

type CheckResult struct {
    Status    string        `json:"status"`
    Latency   time.Duration `json:"latency_ms"`
    Error     string        `json:"error,omitempty"`
    Metadata  interface{}   `json:"metadata,omitempty"`
}

type Checker struct {
    db         DatabasePool
    ftp        FTPClient
    startTime  time.Time
    version    string
}

func (c *Checker) Check(ctx context.Context) HealthCheck {
    checks := make(map[string]CheckResult)

    // Database check
    checks["database"] = c.checkDatabase(ctx)

    // FTP check
    checks["ftp"] = c.checkFTP(ctx)

    // Queue check
    checks["queue"] = c.checkQueue(ctx)

    // Disk space check
    checks["disk"] = c.checkDiskSpace(ctx)

    overall := "healthy"
    for _, check := range checks {
        if check.Status != "healthy" {
            overall = "unhealthy"
            break
        }
    }

    return HealthCheck{
        Status:    overall,
        Version:   c.version,
        Uptime:    time.Since(c.startTime).String(),
        Checks:    checks,
        Timestamp: time.Now(),
    }
}

func (c *Checker) checkDatabase(ctx context.Context) CheckResult {
    start := time.Now()

    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    err := c.db.Ping(ctx)
    latency := time.Since(start)

    if err != nil {
        return CheckResult{
            Status:  "unhealthy",
            Latency: latency,
            Error:   err.Error(),
        }
    }

    // Get pool stats
    stats := c.db.Stats()

    return CheckResult{
        Status:  "healthy",
        Latency: latency,
        Metadata: map[string]interface{}{
            "connections": stats.TotalConns(),
            "idle":        stats.IdleConns(),
        },
    }
}

// API endpoint
GET /api/health
{
  "status": "healthy",
  "version": "1.2.3",
  "uptime": "24h15m30s",
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 5,
      "metadata": {
        "connections": 10,
        "idle": 7
      }
    },
    "ftp": {
      "status": "healthy",
      "latency_ms": 120
    },
    "queue": {
      "status": "healthy",
      "metadata": {
        "pending": 0,
        "processing": 1
      }
    }
  },
  "timestamp": "2024-12-18T10:30:00Z"
}
```

**Priority**: HIGH

---

### 6.6 Docker Image Optimization

**Problem**: Large Docker images, slow builds.

**Solution**:
```dockerfile
# Build stage with cache mounts
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build with cache
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" \
    -o webhook-server ./cmd/webhook-server

# Minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/webhook-server .
COPY --from=builder /build/configs ./configs

# Non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

ENTRYPOINT ["./webhook-server"]

# Result: ~15MB instead of ~800MB
```

**Priority**: MEDIUM

---

### 6.7 Kubernetes Deployment Manifests

**Problem**: Only Docker Compose, no K8s support.

**Solution**:
```yaml
# deployments/k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontol-loader
  labels:
    app: frontol-loader
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontol-loader
  template:
    metadata:
      labels:
        app: frontol-loader
    spec:
      initContainers:
      - name: migrate
        image: frontol-loader:latest
        command: ["/app/migrate", "up"]
        envFrom:
        - secretRef:
            name: frontol-secrets
      containers:
      - name: webhook-server
        image: frontol-loader:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: frontol-config
        - secretRef:
            name: frontol-secrets
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"

---
apiVersion: v1
kind: Service
metadata:
  name: frontol-loader
spec:
  selector:
    app: frontol-loader
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: frontol-loader-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: frontol-loader
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

**Priority**: LOW (if K8s needed)

---

### 6.8 Configuration Hot Reload

**Problem**: Config changes require restart.

**Solution**:
```go
// pkg/config/watcher.go
type ConfigWatcher struct {
    config   *Config
    mu       sync.RWMutex
    onChange []func(*Config)
}

func NewConfigWatcher(path string) (*ConfigWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    cw := &ConfigWatcher{
        config: loadConfig(path),
    }

    go func() {
        for {
            select {
            case event := <-watcher.Events:
                if event.Op&fsnotify.Write == fsnotify.Write {
                    newConfig := loadConfig(path)
                    cw.mu.Lock()
                    cw.config = newConfig
                    cw.mu.Unlock()

                    for _, fn := range cw.onChange {
                        fn(newConfig)
                    }
                }
            }
        }
    }()

    watcher.Add(path)
    return cw, nil
}

func (cw *ConfigWatcher) OnChange(fn func(*Config)) {
    cw.onChange = append(cw.onChange, fn)
}

// Usage
configWatcher.OnChange(func(cfg *Config) {
    logger.SetLevel(cfg.LogLevel)
    rateLimiter.UpdateRate(cfg.RateLimit)
})
```

**Priority**: LOW

---

## 7. Monitoring & Observability

### 7.1 Prometheus Metrics

**Problem**: No metrics exposed.

**Solution**:
```go
// pkg/metrics/prometheus.go
import "github.com/prometheus/client_golang/prometheus"

var (
    etlDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "etl_pipeline_duration_seconds",
            Help:    "ETL pipeline execution duration",
            Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
        },
        []string{"status"},
    )

    filesProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "etl_files_processed_total",
            Help: "Total files processed",
        },
        []string{"kassa_code", "status"},
    )

    transactionsLoaded = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "etl_transactions_loaded_total",
            Help: "Total transactions loaded",
        },
        []string{"table_name"},
    )

    ftpOperationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ftp_operation_duration_seconds",
            Help:    "FTP operation duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation"},
    )

    dbConnectionsInUse = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_in_use",
            Help: "Database connections currently in use",
        },
    )

    queueDepth = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "webhook_queue_depth",
            Help: "Number of items in queue",
        },
        []string{"operation_type"},
    )
)

func init() {
    prometheus.MustRegister(
        etlDuration,
        filesProcessed,
        transactionsLoaded,
        ftpOperationDuration,
        dbConnectionsInUse,
        queueDepth,
    )
}

// Instrumentation
func (p *Pipeline) Run(ctx context.Context, date string) error {
    start := time.Now()

    err := p.run(ctx, date)

    status := "success"
    if err != nil {
        status = "error"
    }

    etlDuration.WithLabelValues(status).Observe(time.Since(start).Seconds())

    return err
}

// Metrics endpoint
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

**Priority**: HIGH

---

### 7.2 Distributed Tracing

**Problem**: Hard to debug issues across pipeline stages.

**Solution**:
```go
// pkg/tracing/tracer.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("frontol-loader")

func (p *Pipeline) Run(ctx context.Context, date string) error {
    ctx, span := tracer.Start(ctx, "pipeline.Run",
        trace.WithAttributes(
            attribute.String("date", date),
        ),
    )
    defer span.End()

    // Clear folders
    ctx, clearSpan := tracer.Start(ctx, "pipeline.ClearFolders")
    err := p.clearFolders(ctx)
    clearSpan.End()
    if err != nil {
        span.RecordError(err)
        return err
    }

    // Send requests
    ctx, sendSpan := tracer.Start(ctx, "pipeline.SendRequests")
    err = p.sendRequests(ctx, date)
    sendSpan.End()
    if err != nil {
        span.RecordError(err)
        return err
    }

    // Process files
    ctx, processSpan := tracer.Start(ctx, "pipeline.ProcessFiles")
    err = p.processFiles(ctx)
    processSpan.End()

    return err
}

// Configure exporter
import "go.opentelemetry.io/otel/exporters/jaeger"

exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
    jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
))

// View traces in Jaeger UI
TRACING_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

**Priority**: MEDIUM

---

### 7.3 Structured Logging Improvements

**Problem**: Log fields inconsistent.

**Solution**:
```go
// pkg/logger/logger.go
type Logger struct {
    *slog.Logger
    fields map[string]interface{}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
    newFields := make(map[string]interface{})
    for k, v := range l.fields {
        newFields[k] = v
    }
    for k, v := range fields {
        newFields[k] = v
    }

    return &Logger{
        Logger: l.Logger,
        fields: newFields,
    }
}

func (l *Logger) WithKassa(code string) *Logger {
    return l.WithFields(map[string]interface{}{
        "kassa_code": code,
    })
}

func (l *Logger) WithFile(name string) *Logger {
    return l.WithFields(map[string]interface{}{
        "file_name": name,
    })
}

// Usage
logger := logger.WithKassa("001").WithFile("export.txt")
logger.Info("Processing file")
// Output: {"level":"info","kassa_code":"001","file_name":"export.txt","msg":"Processing file"}
```

**Priority**: MEDIUM

---

### 7.4 Error Tracking Integration

**Problem**: No centralized error tracking.

**Solution**:
```go
// pkg/errors/sentry.go
import "github.com/getsentry/sentry-go"

func InitSentry(dsn string) error {
    return sentry.Init(sentry.ClientOptions{
        Dsn:              dsn,
        Environment:      os.Getenv("APP_ENV"),
        Release:          os.Getenv("APP_VERSION"),
        AttachStacktrace: true,
        TracesSampleRate: 0.1,
    })
}

func CaptureError(err error, ctx context.Context) {
    hub := sentry.GetHubFromContext(ctx)
    if hub == nil {
        hub = sentry.CurrentHub()
    }

    hub.WithScope(func(scope *sentry.Scope) {
        // Add context
        if kassaCode := ctx.Value("kassa_code"); kassaCode != nil {
            scope.SetTag("kassa_code", kassaCode.(string))
        }

        scope.SetLevel(sentry.LevelError)
        hub.CaptureException(err)
    })
}

// Usage
if err := pipeline.Run(ctx, date); err != nil {
    errors.CaptureError(err, ctx)
    return err
}

// Configuration
SENTRY_DSN=https://...@sentry.io/123
SENTRY_ENVIRONMENT=production
SENTRY_TRACES_SAMPLE_RATE=0.1
```

**Priority**: MEDIUM

---

### 7.5 Performance Profiling Endpoints

**Problem**: Hard to debug performance issues in production.

**Solution**:
```go
// pkg/server/profiling.go
import _ "net/http/pprof"

func (s *Server) setupProfilingRoutes() {
    if !s.cfg.EnableProfiling {
        return
    }

    // CPU profiling
    s.router.HandleFunc("/debug/pprof/profile", pprof.Profile)

    // Heap profiling
    s.router.HandleFunc("/debug/pprof/heap", pprof.Index)

    // Goroutine profiling
    s.router.HandleFunc("/debug/pprof/goroutine", pprof.Index)

    // Block profiling
    runtime.SetBlockProfileRate(1)
    s.router.HandleFunc("/debug/pprof/block", pprof.Index)

    // Mutex profiling
    runtime.SetMutexProfileFraction(1)
    s.router.HandleFunc("/debug/pprof/mutex", pprof.Index)
}

// Usage
// CPU profile: curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
// Heap profile: curl http://localhost:8080/debug/pprof/heap > heap.prof
// Analyze: go tool pprof cpu.prof

// Configuration
ENABLE_PROFILING=true  # Only in staging/dev
PROFILING_AUTH_REQUIRED=true
```

**Priority**: MEDIUM

---

### 7.6 Custom Dashboard

**Problem**: No unified view of system state.

**Solution**:
```go
// pkg/dashboard/dashboard.go
type Dashboard struct {
    Metrics    MetricsSummary
    Health     HealthStatus
    Queue      QueueStatus
    Recent     []RecentRun
    Errors     []ErrorSummary
}

type MetricsSummary struct {
    FilesProcessedToday     int
    TransactionsLoadedToday int
    AvgETLDuration          time.Duration
    SuccessRate             float64
}

// API endpoint
GET /api/dashboard
{
  "metrics": {
    "files_processed_today": 120,
    "transactions_loaded_today": 45000,
    "avg_etl_duration_seconds": 185,
    "success_rate": 0.98
  },
  "health": {
    "status": "healthy",
    "database": "healthy",
    "ftp": "healthy"
  },
  "queue": {
    "pending": 2,
    "processing": 1,
    "completed_today": 48
  },
  "recent_runs": [
    {
      "date": "2024-12-18",
      "status": "success",
      "duration_seconds": 180,
      "files_processed": 5
    }
  ],
  "errors": [
    {
      "count": 3,
      "message": "FTP connection timeout",
      "last_occurrence": "2024-12-18T10:30:00Z"
    }
  ]
}
```

**Priority**: LOW

---

## 8. Error Handling & Resilience

### 8.1 Circuit Breaker for FTP

**Problem**: Repeated FTP failures can cascade.

**Solution**:
```go
// pkg/resilience/circuit_breaker.go
import "github.com/sony/gobreaker"

type CircuitBreaker struct {
    cb *gobreaker.CircuitBreaker
}

func NewCircuitBreaker(name string) *CircuitBreaker {
    return &CircuitBreaker{
        cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
            Name:        name,
            MaxRequests: 3,
            Interval:    time.Minute,
            Timeout:     30 * time.Second,
            ReadyToTrip: func(counts gobreaker.Counts) bool {
                failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
                return counts.Requests >= 3 && failureRatio >= 0.6
            },
            OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
                log.Info("Circuit breaker state changed",
                    "name", name,
                    "from", from,
                    "to", to,
                )
            },
        }),
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    _, err := cb.cb.Execute(func() (interface{}, error) {
        return nil, fn()
    })
    return err
}

// Usage
ftpCircuitBreaker := resilience.NewCircuitBreaker("ftp")

err := ftpCircuitBreaker.Execute(func() error {
    return ftpClient.DownloadFile(file)
})

if err == gobreaker.ErrOpenState {
    // Circuit is open, fail fast
    return errors.New("FTP service unavailable")
}
```

**Priority**: MEDIUM

---

### 8.2 Retry with Exponential Backoff

**Problem**: Retry logic duplicated across codebase.

**Solution**:
```go
// pkg/resilience/retry.go
type RetryPolicy struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func (rp *RetryPolicy) Execute(ctx context.Context, fn func() error) error {
    var lastErr error
    delay := rp.InitialDelay

    for attempt := 0; attempt < rp.MaxAttempts; attempt++ {
        if attempt > 0 {
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return ctx.Err()
            }

            delay = time.Duration(float64(delay) * rp.Multiplier)
            if delay > rp.MaxDelay {
                delay = rp.MaxDelay
            }
        }

        err := fn()
        if err == nil {
            return nil
        }

        // Don't retry on certain errors
        if !isRetryable(err) {
            return err
        }

        lastErr = err
        log.Warn("Retrying after error",
            "attempt", attempt+1,
            "max_attempts", rp.MaxAttempts,
            "delay", delay,
            "error", err,
        )
    }

    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Usage
retry := &resilience.RetryPolicy{
    MaxAttempts:  5,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     5 * time.Second,
    Multiplier:   2.0,
}

err := retry.Execute(ctx, func() error {
    return ftpClient.DownloadFile(file)
})
```

**Priority**: MEDIUM

---

### 8.3 Timeout Management

**Problem**: No timeouts on FTP operations.

**Solution**:
```go
// pkg/ftp/ftp.go
func (c *Client) DownloadFileWithTimeout(ctx context.Context, remotePath string, timeout time.Duration) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    type result struct {
        data []byte
        err  error
    }

    resultChan := make(chan result, 1)

    go func() {
        data, err := c.downloadFile(remotePath)
        resultChan <- result{data, err}
    }()

    select {
    case res := <-resultChan:
        return res.data, res.err
    case <-ctx.Done():
        return nil, fmt.Errorf("download timeout: %w", ctx.Err())
    }
}

// Configuration
FTP_DOWNLOAD_TIMEOUT_SECONDS=300  # 5 minutes
FTP_LIST_TIMEOUT_SECONDS=30
FTP_UPLOAD_TIMEOUT_SECONDS=60
```

**Priority**: HIGH

---

### 8.4 Dead Letter Queue

**Problem**: Failed ETL runs lost.

**Solution**:
```go
// pkg/queue/dlq.go
type DeadLetterQueue struct {
    store Storage
}

func (dlq *DeadLetterQueue) Add(ctx context.Context, item FailedItem) error {
    return dlq.store.Save(ctx, item)
}

type FailedItem struct {
    RequestID   string
    Date        string
    Error       string
    Attempts    int
    FirstFailed time.Time
    LastFailed  time.Time
    Metadata    map[string]interface{}
}

// Storage in database
CREATE TABLE dead_letter_queue (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID NOT NULL,
    operation_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    error TEXT NOT NULL,
    attempts INT DEFAULT 1,
    first_failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB
);

// Retry mechanism
func (dlq *DeadLetterQueue) RetryAll(ctx context.Context) error {
    items, err := dlq.store.GetAll(ctx)
    if err != nil {
        return err
    }

    for _, item := range items {
        if item.Attempts >= 3 {
            continue // Max retries reached
        }

        // Re-queue
        queue.Enqueue(ctx, item.Payload)
        dlq.store.Delete(ctx, item.ID)
    }

    return nil
}

// API endpoints
GET  /api/dlq        # List failed items
POST /api/dlq/retry  # Retry all
```

**Priority**: MEDIUM

---

### 8.5 Partial Failure Recovery

**Problem**: Pipeline fails completely on any error.

**Solution**:
```go
// pkg/pipeline/recovery.go
type ProcessingResult struct {
    Succeeded []FileResult
    Failed    []FileError
}

type FileResult struct {
    File          string
    Transactions  int
    Duration      time.Duration
}

type FileError struct {
    File  string
    Error error
}

func (p *Pipeline) ProcessFilesWithRecovery(ctx context.Context, tasks []FileTask) ProcessingResult {
    var mu sync.Mutex
    result := ProcessingResult{
        Succeeded: make([]FileResult, 0),
        Failed:    make([]FileError, 0),
    }

    var wg sync.WaitGroup
    for _, task := range tasks {
        wg.Add(1)
        go func(t FileTask) {
            defer wg.Done()

            start := time.Now()
            count, err := p.processFile(ctx, t)

            mu.Lock()
            defer mu.Unlock()

            if err != nil {
                result.Failed = append(result.Failed, FileError{
                    File:  t.FileName,
                    Error: err,
                })
            } else {
                result.Succeeded = append(result.Succeeded, FileResult{
                    File:         t.FileName,
                    Transactions: count,
                    Duration:     time.Since(start),
                })
            }
        }(task)
    }

    wg.Wait()
    return result
}

// Return detailed report
{
  "succeeded": 45,
  "failed": 2,
  "failed_files": [
    {"file": "kassa001/sales.txt", "error": "parse error at line 150"},
    {"file": "kassa002/sales.txt", "error": "database timeout"}
  ]
}
```

**Priority**: HIGH

---

## 9. Developer Experience

### 9.1 CLI Improvements

**Problem**: Limited CLI functionality.

**Solution**:
```go
// cmd/frontol-cli/main.go
import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "frontol",
    Short: "Frontol ETL CLI",
}

var loadCmd = &cobra.Command{
    Use:   "load [date]",
    Short: "Load data for specific date",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // ... load logic
    },
}

var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Validate configuration",
    Run: func(cmd *cobra.Command, args []string) {
        cfg, err := config.LoadConfig()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        if err := cfg.Validate(); err != nil {
            fmt.Fprintf(os.Stderr, "Validation failed:\n%v\n", err)
            os.Exit(1)
        }

        fmt.Println("✓ Configuration valid")
    },
}

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show pipeline status",
    Run: func(cmd *cobra.Command, args []string) {
        // Query webhook server for status
    },
}

// Commands
frontol load 2024-12-18
frontol load --range=2024-12-01:2024-12-31
frontol validate
frontol status
frontol retry --request-id=abc123
frontol stats --date=2024-12-18
```

**Priority**: MEDIUM

---

### 9.2 Development Environment Setup

**Problem**: Complex setup for new developers.

**Solution**:
```bash
# scripts/dev-setup.sh
#!/bin/bash

set -e

echo "Setting up Frontol ETL development environment..."

# Install dependencies
echo "Installing Go dependencies..."
go mod download

# Install tools
echo "Installing development tools..."
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0
go install github.com/cosmtrek/air@latest  # Live reload
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Start services
echo "Starting Docker services..."
docker-compose -f docker-compose.dev.yml up -d

# Wait for PostgreSQL
echo "Waiting for PostgreSQL..."
until docker-compose -f docker-compose.dev.yml exec -T postgres pg_isready; do
    sleep 1
done

# Run migrations
echo "Running migrations..."
go run ./cmd/migrate/main.go up

# Seed test data
echo "Seeding test data..."
go run ./scripts/seed-test-data.go

# Configure git hooks
echo "Setting up git hooks..."
cp scripts/hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit

echo "✓ Development environment ready!"
echo ""
echo "Commands:"
echo "  make dev       # Start with live reload"
echo "  make test      # Run tests"
echo "  make lint      # Run linter"
```

**Priority**: MEDIUM

---

### 9.3 Makefile Enhancements

**Problem**: Limited make targets.

**Solution**:
```makefile
.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: dev
dev: ## Start development server with live reload
	air

.PHONY: test-watch
test-watch: ## Run tests in watch mode
	gotestsum --watch -- -v ./...

.PHONY: coverage
coverage: ## Generate coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

.PHONY: bench
bench: ## Run benchmarks
	go test -bench=. -benchmem -cpuprofile=cpu.prof ./...
	go tool pprof -http=:8080 cpu.prof

.PHONY: generate
generate: ## Generate code (mocks, etc.)
	go generate ./...

.PHONY: install-tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0
	go install gotest.tools/gotestsum@latest
	go install github.com/cosmtrek/air@latest

.PHONY: db-reset
db-reset: ## Reset database
	go run ./cmd/migrate/main.go down
	go run ./cmd/migrate/main.go up
	go run ./scripts/seed-test-data.go

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	docker-compose down -v
	docker system prune -f

.PHONY: security-scan
security-scan: ## Run security scan
	gosec ./...
	trivy fs .
```

**Priority**: LOW

---

### 9.4 Code Generation for Transaction Types

**Problem**: Boilerplate code for each transaction type.

**Solution**:
```go
//go:generate go run scripts/generate-transaction.go -type=NewTransactionType -table=new_transactions

// scripts/generate-transaction.go
// Generates:
// - pkg/models/new_transaction.go
// - pkg/parser/new_transaction_parser.go
// - pkg/repository/new_transaction_loader.go
// - migrations/NNNNNN_create_new_transactions_table.up.sql
// - tests for all above

// Usage
go generate ./...
```

**Priority**: MEDIUM

---

### 9.5 Interactive Debugger Setup

**Problem**: Debugging requires manual breakpoint setting.

**Solution**:
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Webhook Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/webhook-server",
      "env": {
        "DB_PASSWORD": "test",
        "FTP_PASSWORD": "test"
      },
      "envFile": "${workspaceFolder}/configs/development.env"
    },
    {
      "name": "Debug ETL Loader",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/loader",
      "args": ["2024-12-18"]
    },
    {
      "name": "Attach to Running Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": "${command:pickProcess}"
    }
  ]
}
```

**Priority**: LOW

---

## 10. New Features & Capabilities

### 10.1 Incremental Processing

**Problem**: Full reload for entire date.

**Solution**:
```go
// pkg/pipeline/incremental.go
type IncrementalProcessor struct {
    checkpoint CheckpointStore
}

func (p *IncrementalProcessor) ProcessIncremental(ctx context.Context, date string) error {
    lastCheckpoint, err := p.checkpoint.Get(ctx, date)
    if err != nil {
        return err
    }

    files, err := p.listFiles(ctx)
    if err != nil {
        return err
    }

    // Only process files modified after checkpoint
    newFiles := filterFilesAfter(files, lastCheckpoint.Timestamp)

    for _, file := range newFiles {
        if err := p.processFile(ctx, file); err != nil {
            return err
        }

        // Update checkpoint
        p.checkpoint.Set(ctx, date, Checkpoint{
            Timestamp: time.Now(),
            LastFile:  file.Name,
        })
    }

    return nil
}

// Database
CREATE TABLE processing_checkpoints (
    date DATE PRIMARY KEY,
    last_processed_at TIMESTAMPTZ NOT NULL,
    last_file TEXT,
    metadata JSONB
);
```

**Priority**: MEDIUM

---

### 10.2 Data Validation Rules Engine

**Problem**: Hardcoded validation logic.

**Solution**:
```go
// pkg/validation/rules.go
type Rule interface {
    Name() string
    Validate(ctx context.Context, data interface{}) error
}

type RuleEngine struct {
    rules map[string][]Rule
}

func (re *RuleEngine) Register(transactionType string, rule Rule) {
    re.rules[transactionType] = append(re.rules[transactionType], rule)
}

func (re *RuleEngine) Validate(ctx context.Context, txType string, data interface{}) error {
    rules := re.rules[txType]
    var errs []error

    for _, rule := range rules {
        if err := rule.Validate(ctx, data); err != nil {
            errs = append(errs, fmt.Errorf("%s: %w", rule.Name(), err))
        }
    }

    if len(errs) > 0 {
        return joinErrors(errs)
    }

    return nil
}

// Example rules
type AmountRangeRule struct {
    Min, Max float64
}

func (r AmountRangeRule) Validate(ctx context.Context, data interface{}) error {
    tx := data.(*models.Transaction)
    if tx.Amount < r.Min || tx.Amount > r.Max {
        return fmt.Errorf("amount %.2f outside range [%.2f, %.2f]", tx.Amount, r.Min, r.Max)
    }
    return nil
}

// Configuration-driven rules
{
  "validation_rules": {
    "transaction_type_1": [
      {"type": "amount_range", "min": 0, "max": 100000},
      {"type": "required_fields", "fields": ["customer_id", "amount"]},
      {"type": "date_not_future"}
    ]
  }
}
```

**Priority**: MEDIUM

---

### 10.3 Data Export API

**Problem**: No way to export processed data.

**Solution**:
```go
// pkg/exporter/exporter.go
type Exporter interface {
    Export(ctx context.Context, query ExportQuery) (io.Reader, error)
}

type ExportQuery struct {
    DateFrom        time.Time
    DateTo          time.Time
    KassaCodes      []string
    TransactionType string
    Format          string // csv, json, excel
}

type CSVExporter struct {
    db DatabasePool
}

func (e *CSVExporter) Export(ctx context.Context, q ExportQuery) (io.Reader, error) {
    rows, err := e.db.Query(ctx, buildQuery(q))
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)

    // Write header
    writer.Write(getColumnNames(q.TransactionType))

    // Write rows
    for rows.Next() {
        var record []string
        // ... scan and write
        writer.Write(record)
    }

    writer.Flush()
    return &buf, nil
}

// API endpoint
GET /api/export?date_from=2024-12-01&date_to=2024-12-31&format=csv&kassa=001,002
```

**Priority**: LOW

---

### 10.4 Scheduled ETL Jobs

**Problem**: Manual triggering only.

**Solution**:
```go
// pkg/scheduler/scheduler.go
import "github.com/robfig/cron/v3"

type Scheduler struct {
    cron   *cron.Cron
    runner PipelineRunner
}

func NewScheduler(runner PipelineRunner) *Scheduler {
    return &Scheduler{
        cron:   cron.New(),
        runner: runner,
    }
}

func (s *Scheduler) Start(ctx context.Context) error {
    // Daily at 2 AM
    s.cron.AddFunc("0 2 * * *", func() {
        yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
        s.runner.Run(ctx, yesterday)
    })

    // Configurable schedules
    schedules := loadSchedulesFromConfig()
    for _, sched := range schedules {
        s.cron.AddFunc(sched.CronExpr, func() {
            s.runner.Run(ctx, sched.Date)
        })
    }

    s.cron.Start()

    <-ctx.Done()
    s.cron.Stop()
    return nil
}

// Configuration
SCHEDULE_DAILY_ETL=0 2 * * *
SCHEDULE_HOURLY_CHECK=0 * * * *
```

**Priority**: LOW

---

### 10.5 Multi-Tenancy Support

**Problem**: Single tenant only.

**Solution**:
```go
// pkg/tenant/tenant.go
type Tenant struct {
    ID     string
    Name   string
    Config TenantConfig
}

type TenantConfig struct {
    FTPConfig      FTPConfig
    DatabaseDSN    string
    KassaStructure string
}

type TenantManager struct {
    tenants map[string]*Tenant
}

func (tm *TenantManager) GetTenant(id string) (*Tenant, error) {
    tenant, ok := tm.tenants[id]
    if !ok {
        return nil, errors.New("tenant not found")
    }
    return tenant, nil
}

// Middleware
func TenantMiddleware(tm *TenantManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantID := r.Header.Get("X-Tenant-ID")
            tenant, err := tm.GetTenant(tenantID)
            if err != nil {
                http.Error(w, "Invalid tenant", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), "tenant", tenant)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Database schema
CREATE SCHEMA tenant_001;
CREATE SCHEMA tenant_002;

-- Switch schema per request
SET search_path TO tenant_001;
```

**Priority**: LOW (if needed)

---

### 10.6 Real-time Notifications

**Problem**: No notifications on completion/errors.

**Solution**:
```go
// pkg/notifier/notifier.go
type Notifier interface {
    Notify(ctx context.Context, event Event) error
}

type Event struct {
    Type      string
    Severity  string
    Message   string
    Metadata  map[string]interface{}
    Timestamp time.Time
}

type SlackNotifier struct {
    webhookURL string
}

func (s *SlackNotifier) Notify(ctx context.Context, event Event) error {
    payload := map[string]interface{}{
        "text": event.Message,
        "attachments": []map[string]interface{}{
            {
                "color": getSeverityColor(event.Severity),
                "fields": []map[string]interface{}{
                    {"title": "Type", "value": event.Type, "short": true},
                    {"title": "Time", "value": event.Timestamp.Format(time.RFC3339), "short": true},
                },
            },
        },
    }

    return sendToSlack(s.webhookURL, payload)
}

type EmailNotifier struct {
    smtp SMTPConfig
}

type MultiNotifier struct {
    notifiers []Notifier
}

func (m *MultiNotifier) Notify(ctx context.Context, event Event) error {
    var errs []error
    for _, n := range m.notifiers {
        if err := n.Notify(ctx, event); err != nil {
            errs = append(errs, err)
        }
    }
    return joinErrors(errs)
}

// Usage
notifier.Notify(ctx, Event{
    Type:     "etl.completed",
    Severity: "info",
    Message:  "ETL completed successfully for 2024-12-18",
    Metadata: map[string]interface{}{
        "files_processed": 45,
        "duration":        "3m15s",
    },
})

// Configuration
NOTIFIER_SLACK_WEBHOOK=https://hooks.slack.com/...
NOTIFIER_EMAIL_TO=team@example.com
NOTIFIER_TELEGRAM_BOT_TOKEN=...
```

**Priority**: LOW

---

## Summary: Priority Matrix

### Immediate (Sprint 1)
1. Fix race condition in queue workers
2. Remove duplicate processFile code
3. Constant-time token comparison
4. Context propagation fix
5. Input validation framework
6. Graceful shutdown improvements
7. Health check enhancements
8. FTP connection pooling
9. Worker pool for file processing
10. Batch FTP listing
11. Operation type caching
12. Integration test framework
13. Prometheus metrics
14. Configuration validation

### Short-term (Sprint 2-3)
15. Plugin architecture for transactions
16. Repository pattern with decorators
17. Service layer separation
18. Rate limiting
19. Audit logging
20. COPY for bulk loads
21. Timeout management
22. Partial failure recovery
23. Test data builders
24. Distributed tracing
25. Error tracking (Sentry)

### Medium-term (Quarter)
26. Event-driven architecture
27. Strategy pattern for processing
28. Circuit breaker
29. Retry framework
30. Dead letter queue
31. Multi-environment config
32. Feature flags
33. Performance profiling endpoints
34. Benchmark suite
35. CLI improvements

### Long-term (Roadmap)
36. Incremental processing
37. Data validation rules engine
38. Data export API
39. Scheduled jobs
40. Multi-tenancy (if needed)
41. Real-time notifications
42. Kubernetes manifests (if needed)

---

**Total Improvements**: 42 items across 10 categories
**Estimated Effort**: 6-8 months for full implementation
**Quick Wins** (< 2 days each): Items 1, 2, 3, 4, 7, 9, 13, 14

This plan enables painless feature additions by establishing:
- Clear architectural boundaries
- Plugin-based extensibility
- Comprehensive testing infrastructure
- Production-ready observability
- Robust error handling
