//go:build integration
// +build integration

package framework

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/migrate"
	"github.com/user/go-frontol-loader/pkg/models"
)

const (
	postgresImage    = "postgres:16-alpine"
	postgresUser     = "frontol_user"
	postgresPassword = "test_password"
	postgresDB       = "kassa_db_test"
)

// PostgresContainer wraps testcontainers PostgreSQL instance
type PostgresContainer struct {
	Container testcontainers.Container
	Config    *models.Config
	Pool      *db.Pool
}

// NewPostgresContainer creates and starts a PostgreSQL test container
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     postgresUser,
			"POSTGRES_PASSWORD": postgresPassword,
			"POSTGRES_DB":       postgresDB,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	cfg := &models.Config{
		DBHost:     host,
		DBPort:     mappedPort.Int(),
		DBUser:     postgresUser,
		DBPassword: postgresPassword,
		DBName:     postgresDB,
		DBSSLMode:  "disable",
	}

	// Create connection pool
	pool, err := db.NewPool(cfg)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	return &PostgresContainer{
		Container: container,
		Config:    cfg,
		Pool:      pool,
	}, nil
}

// Close closes the database pool and terminates the container
func (pc *PostgresContainer) Close(ctx context.Context) error {
	if pc.Pool != nil {
		pc.Pool.Close()
	}
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}

// RunMigrations runs database migrations
func (pc *PostgresContainer) RunMigrations(ctx context.Context) error {
	// Get migrations path relative to project root
	migrationsPath := filepath.Join("..", "..", "..", "pkg", "migrate", "migrations")

	migrator, err := migrate.NewMigratorFromPath(pc.Config, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Truncate truncates all tables for clean test state
func (pc *PostgresContainer) Truncate(ctx context.Context) error {
	tables := make([]string, 0, len(models.TxSchemas))
	for table := range models.TxSchemas {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		_, err := pc.Pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// GetDSN returns the database connection string
func (pc *PostgresContainer) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pc.Config.DBUser,
		pc.Config.DBPassword,
		pc.Config.DBHost,
		pc.Config.DBPort,
		pc.Config.DBName,
		pc.Config.DBSSLMode,
	)
}
