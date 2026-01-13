// Package migrate provides database migration utilities using golang-migrate.
package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver

	"github.com/user/go-frontol-loader/pkg/models"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrator handles database migrations
type Migrator struct {
	db      *sql.DB
	migrate *migrate.Migrate
	dbName  string
}

// NewMigrator creates a new Migrator instance
func NewMigrator(cfg *models.Config) (*Migrator, error) {
	// First, ensure database exists
	if err := ensureDatabaseExists(cfg); err != nil {
		return nil, fmt.Errorf("failed to ensure database exists: %w", err)
	}

	// Build connection string for database/sql
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	// Open database connection
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("Failed to close database after ping error",
				"error", closeErr.Error(),
			)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("Failed to close database after driver error",
				"error", closeErr.Error(),
			)
		}
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migration source from embedded filesystem
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("Failed to close database after source error",
				"error", closeErr.Error(),
			)
		}
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, cfg.DBName, driver)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("Failed to close database after migrate init error",
				"error", closeErr.Error(),
			)
		}
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		db:      db,
		migrate: m,
		dbName:  cfg.DBName,
	}, nil
}

// NewMigratorFromPath creates a Migrator using migration files from filesystem path
func NewMigratorFromPath(cfg *models.Config, migrationsPath string) (*Migrator, error) {
	// Build connection string
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	// Create migrate instance from path
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate: m,
		dbName:  cfg.DBName,
	}, nil
}

// Close closes the migrator and releases resources
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Down rolls back all migrations
func (m *Migrator) Down() error {
	err := m.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	return nil
}

// Steps runs n migrations (positive = up, negative = down)
func (m *Migrator) Steps(n int) error {
	err := m.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration steps: %w", err)
	}
	return nil
}

// Version returns current migration version
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Force sets migration version without running migrations
func (m *Migrator) Force(version int) error {
	return m.migrate.Force(version)
}

// Drop drops all tables in the database
func (m *Migrator) Drop() error {
	return m.migrate.Drop()
}

// Status returns migration status information
type Status struct {
	Version uint
	Dirty   bool
	Error   error
}

// GetStatus returns current migration status
func (m *Migrator) GetStatus() Status {
	version, dirty, err := m.migrate.Version()
	return Status{
		Version: version,
		Dirty:   dirty,
		Error:   err,
	}
}

// ensureDatabaseExists creates database if it doesn't exist
func ensureDatabaseExists(cfg *models.Config) error {
	// Connect to postgres database to check/create target database
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBSSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("Failed to close database",
				"error", err.Error(),
			)
		}
	}()

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, cfg.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		slog.Info("Creating database",
			"database", cfg.DBName,
			"event", "db_create",
		)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		slog.Info("Database created successfully",
			"database", cfg.DBName,
			"event", "db_created",
		)
	}

	return nil
}
