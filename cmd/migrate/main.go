// Command migrate provides database migration management.
// Usage:
//
//	migrate up              - Apply all pending migrations
//	migrate down            - Rollback all migrations
//	migrate step N          - Apply N migrations (negative to rollback)
//	migrate version         - Show current migration version
//	migrate force VERSION   - Force set migration version
//	migrate drop            - Drop all tables (DANGEROUS!)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/migrate"
)

func main() {
	// Parse flags
	migrationsPath := flag.String("path", "", "Path to migrations directory (optional, uses embedded by default)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// Load database configuration only (migrations don't need FTP config)
	cfg, err := config.LoadDBConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create migrator
	var migrator *migrate.Migrator
	if *migrationsPath != "" {
		migrator, err = migrate.NewMigratorFromPath(cfg, *migrationsPath)
	} else {
		migrator, err = migrate.NewMigrator(cfg)
	}
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			log.Printf("Failed to close migrator: %v", err)
		}
	}()

	// Execute command
	switch command {
	case "up":
		fmt.Println("Applying all pending migrations...")
		if err := migrator.Up(); err != nil {
			closeMigrator(migrator)
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✓ Migrations applied successfully")

	case "down":
		fmt.Println("Rolling back all migrations...")
		if err := migrator.Down(); err != nil {
			closeMigrator(migrator)
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✓ Migrations rolled back successfully")

	case "step":
		if len(args) < 2 {
			closeMigrator(migrator)
			log.Fatal("step command requires a number argument")
		}
		var n int
		if _, err := fmt.Sscanf(args[1], "%d", &n); err != nil {
			closeMigrator(migrator)
			log.Fatalf("Invalid step number: %s", args[1])
		}
		fmt.Printf("Running %d migration step(s)...\n", n)
		if err := migrator.Steps(n); err != nil {
			closeMigrator(migrator)
			log.Fatalf("Migration step failed: %v", err)
		}
		fmt.Println("✓ Migration step completed successfully")

	case "version":
		status := migrator.GetStatus()
		if status.Error != nil {
			fmt.Printf("No migrations applied yet or error: %v\n", status.Error)
		} else {
			fmt.Printf("Current version: %d\n", status.Version)
			if status.Dirty {
				fmt.Println("⚠ Database is in dirty state!")
			}
		}

	case "force":
		if len(args) < 2 {
			closeMigrator(migrator)
			log.Fatal("force command requires a version argument")
		}
		var version int
		if _, err := fmt.Sscanf(args[1], "%d", &version); err != nil {
			closeMigrator(migrator)
			log.Fatalf("Invalid version: %s", args[1])
		}
		fmt.Printf("Forcing version to %d...\n", version)
		if err := migrator.Force(version); err != nil {
			closeMigrator(migrator)
			log.Fatalf("Force version failed: %v", err)
		}
		fmt.Println("✓ Version set successfully")

	case "drop":
		fmt.Println("⚠ WARNING: This will drop all tables!")
		fmt.Print("Type 'yes' to confirm: ")
		var confirm string
		_, _ = fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("Aborted")
			os.Exit(0)
		}
		fmt.Println("Dropping all tables...")
		if err := migrator.Drop(); err != nil {
			closeMigrator(migrator)
			log.Fatalf("Drop failed: %v", err)
		}
		fmt.Println("✓ All tables dropped")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: migrate [flags] <command> [args]

Commands:
  up              Apply all pending migrations
  down            Rollback all migrations
  step N          Apply N migrations (negative to rollback)
  version         Show current migration version
  force VERSION   Force set migration version (use when dirty)
  drop            Drop all tables (DANGEROUS!)

Flags:
  -path string    Path to migrations directory (optional)

Examples:
  migrate up                    # Apply all migrations
  migrate down                  # Rollback all migrations
  migrate step 1                # Apply 1 migration
  migrate step -1               # Rollback 1 migration
  migrate version               # Show current version
  migrate force 2               # Force version to 2
  migrate -path ./migrations up # Use migrations from path`)
}

func closeMigrator(migrator *migrate.Migrator) {
	if err := migrator.Close(); err != nil {
		log.Printf("Failed to close migrator: %v", err)
	}
}
