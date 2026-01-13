# Integration Test Framework

Testcontainers-based framework for integration testing with automated PostgreSQL and FTP setup.

## Overview

This framework provides:
- **Automated container management**: PostgreSQL and FTP containers via testcontainers-go
- **Database migrations**: Automatic migration execution
- **Fixture helpers**: Easy test data creation
- **Seed data**: Predefined datasets for common scenarios
- **Clean state management**: Automatic cleanup and reset capabilities

## Quick Start

### Basic Test Setup

```go
func TestMyFeature(t *testing.T) {
    // Setup test environment (PostgreSQL + FTP containers)
    env := framework.SetupTestEnvironment(t)

    // Reset to clean state
    env.Reset(t)

    // Your test code here
    ctx := env.GetContext()
    // ... use env.Postgres.Pool, env.FTP.Client, env.Builder
}
```

### With Seed Data

```go
func TestWithData(t *testing.T) {
    env := framework.SetupTestEnvironment(t)
    env.Reset(t)
    env.LoadBasicData(t) // Loads predefined dataset

    // Test with pre-populated data
}
```

## Components

### TestEnvironment

Main entry point for all integration tests. Provides:
- `Postgres`: PostgreSQL container with connection pool
- `FTP`: FTP container with client
- `Builder`: Test data builder for fixtures
- `GetContext()`: Test context
- `Reset(t)`: Clean all data
- `LoadBasicData(t)`: Load predefined seed data
- `Teardown()`: Cleanup (automatic via t.Cleanup)

### PostgresContainer

Manages PostgreSQL test container:
- Automatic migrations
- Connection pooling
- Table truncation
- DSN generation

**Methods:**
```go
container.RunMigrations(ctx)  // Run database migrations
container.Truncate(ctx)       // Clear all tables
container.GetDSN()            // Get connection string
container.Close(ctx)          // Cleanup
```

### FTPContainer

Manages FTP test container:
- Folder structure setup
- File upload/download
- Cleanup operations

**Methods:**
```go
container.SetupFolderStructure(ctx)  // Create kassa folders
container.CleanFolders(ctx)          // Remove all files
container.GetConnectionString()      // Get FTP URL
container.Close(ctx)                 // Cleanup
```

### TestDataBuilder

Helper for creating test data:

**Database operations:**
```go
builder.CreateTransactionRegistration(ctx, transaction)
builder.CountTransactions(ctx, tableName)
builder.GetTransaction(ctx, id, sourceFolder)
```

**FTP operations:**
```go
builder.CreateFTPFile(ctx, kassaCode, folder, filename, content)
```

**Seed data:**
```go
dataset := framework.GetBasicDataSet()
builder.LoadSeedData(ctx, dataset)
```

## Seed Data Sets

### Basic Dataset

Includes:
- 2 transaction registrations
- 1 test file on FTP

```go
dataset := framework.GetBasicDataSet()
```

### Custom Dataset

```go
dataset := framework.SeedDataSet{
    Name: "custom",
    Transactions: []models.TransactionRegistration{
        {
            TransactionIDUnique: 1,
            SourceFolder: "001/folder1",
            // ... other fields
        },
    },
    Files: []framework.TestFile{
        {
            KassaCode: "001",
            FolderName: "folder1",
            Filename: "data.txt",
            Content: "...",
        },
    },
}
builder.LoadSeedData(ctx, dataset)
```

## Examples

### Test Database Operations

```go
func TestDatabaseInsert(t *testing.T) {
    env := framework.SetupTestEnvironment(t)
    env.Reset(t)
    ctx := env.GetContext()

    tr := &models.TransactionRegistration{
        TransactionIDUnique: 123,
        SourceFolder: "001/folder1",
        ItemCode: "TEST001",
        AmountTotal: 100.50,
    }

    err := env.Builder.CreateTransactionRegistration(ctx, tr)
    if err != nil {
        t.Fatalf("Failed to create: %v", err)
    }

    count, _ := env.Builder.CountTransactions(ctx, "transaction_registrations")
    assert.Equal(t, 1, count)
}
```

### Test FTP Operations

```go
func TestFTPUpload(t *testing.T) {
    env := framework.SetupTestEnvironment(t)
    env.Reset(t)
    ctx := env.GetContext()

    content := `#
DB001
REPORT001
1;2024-12-18;10:30:00;1;...`

    err := env.Builder.CreateFTPFile(ctx, "001", "folder1", "test.txt", content)
    assert.NoError(t, err)

    files, _ := env.FTP.Client.ListFiles("/response/001/folder1")
    assert.Contains(t, files, "test.txt")
}
```

### Test ETL Pipeline

```go
func TestETLPipeline(t *testing.T) {
    env := framework.SetupTestEnvironment(t)
    env.Reset(t)

    // Create test file on FTP
    env.Builder.CreateFTPFile(ctx, "001", "folder1", "data.txt", testData)

    // Run pipeline
    result, err := pipeline.Run(ctx, logger, env.Postgres.Config, "2024-12-18")
    assert.NoError(t, err)
    assert.True(t, result.Success)

    // Verify data was loaded
    count, _ := env.Builder.CountTransactions(ctx, "transaction_registrations")
    assert.Greater(t, count, 0)
}
```

## Running Tests

### All Integration Tests

```bash
go test -v -tags=integration ./tests/integration/...
```

### Specific Test

```bash
go test -v -tags=integration ./tests/integration/ -run TestFramework_BasicSetup
```

### Skip Integration Tests

```bash
SKIP_INTEGRATION_TESTS=true go test -v -tags=integration ./tests/integration/...
```

## Requirements

- Docker (for testcontainers)
- Go 1.24+
- ~500MB disk space for container images

## See Also

- `tests/integration/README.md` - общий сценарий запуска интеграционных тестов

## Container Images

- PostgreSQL: `postgres:16-alpine`
- FTP: `fauria/vsftpd:latest`

Images are automatically pulled on first run.

## Performance Tips

1. **Reuse containers**: Use `env.Reset(t)` instead of creating new environment
2. **Parallel tests**: Enable with `t.Parallel()` (containers are isolated)
3. **Cleanup**: Automatic via `t.Cleanup()`, no manual cleanup needed

## Troubleshooting

### Container start fails

Ensure Docker is running:
```bash
docker ps
```

### Port conflicts

Testcontainers uses random ports, but check for:
```bash
docker ps | grep postgres
docker ps | grep vsftpd
```

### Migration errors

Check migration files:
```bash
ls -la pkg/migrate/migrations/
```

### FTP connection timeout

FTP container may need extra time on first start. Framework includes 2s delay.

## Architecture

```
TestEnvironment
├── PostgresContainer
│   ├── Container (testcontainers)
│   ├── Pool (pgx connection pool)
│   └── Config
├── FTPContainer
│   ├── Container (testcontainers)
│   ├── Client (FTP client)
│   └── Config
└── TestDataBuilder
    ├── Database helpers
    ├── FTP helpers
    └── Seed data loaders
```

## Best Practices

1. **Always call Reset**: `env.Reset(t)` for clean state
2. **Use t.Helper()**: Mark helper functions with `t.Helper()`
3. **Cleanup automatic**: Don't manually close containers
4. **Seed data reuse**: Create reusable datasets
5. **Test isolation**: Each test should be independent

## Contributing

When adding new transaction types or features:

1. Update seed datasets in `fixtures.go`
2. Add table names to `Truncate()` in `postgres.go`
3. Document new fixtures in this README
4. Add example tests in `framework_test.go`
