package models

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTxSchemasCoverExactlyMigrationTables(t *testing.T) {
	tables := loadTxTablesFromMigration(t)

	migrationTables := make([]string, 0, len(tables))
	for table := range tables {
		migrationTables = append(migrationTables, table)
	}
	slices.Sort(migrationTables)

	schemaTables := make([]string, 0, len(TxSchemas))
	for table := range TxSchemas {
		schemaTables = append(schemaTables, table)
	}
	slices.Sort(schemaTables)

	if len(schemaTables) != len(migrationTables) {
		t.Fatalf("TxSchemas table count = %d, want %d", len(schemaTables), len(migrationTables))
	}

	for i, table := range migrationTables {
		if schemaTables[i] != table {
			t.Fatalf("TxSchemas tables differ from migration: got %v, want %v", schemaTables, migrationTables)
		}
	}
}

func TestTxSchemasMatchMigrationColumnOrder(t *testing.T) {
	tables := loadTxTablesFromMigration(t)

	for table, schema := range TxSchemas {
		migrationColumns, ok := tables[table]
		if !ok {
			t.Fatalf("table %s missing from migration", table)
		}

		schemaColumns := make([]string, 0, len(schema))
		for _, column := range schema {
			schemaColumns = append(schemaColumns, column.Name)
		}

		if len(schemaColumns) != len(migrationColumns) {
			t.Fatalf("table %s column count = %d, want %d", table, len(schemaColumns), len(migrationColumns))
		}

		for i, column := range migrationColumns {
			if schemaColumns[i] != column {
				t.Fatalf("table %s column %d = %s, want %s", table, i, schemaColumns[i], column)
			}
		}
	}
}

func loadTxTablesFromMigration(t *testing.T) map[string][]string {
	t.Helper()

	root := findRepoRoot(t)
	path := filepath.Join(root, "pkg", "migrate", "migrations", "000001_init_schema.up.sql")

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open migration: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	tables := make(map[string][]string)
	likeTables := make(map[string]string)

	var currentTable string
	var currentColumns []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if strings.HasPrefix(line, "CREATE TABLE ") {
			parts := strings.Fields(line)
			if len(parts) < 3 {
				t.Fatalf("unexpected CREATE TABLE line: %s", line)
			}
			currentTable = parts[2]
			currentColumns = nil
			continue
		}

		if currentTable == "" {
			continue
		}

		switch {
		case line == ");":
			if strings.HasPrefix(currentTable, "tx_") {
				tables[currentTable] = append([]string(nil), currentColumns...)
			}
			currentTable = ""
			currentColumns = nil
		case strings.HasPrefix(line, "LIKE "):
			parts := strings.Fields(line)
			if len(parts) < 2 {
				t.Fatalf("unexpected LIKE line: %s", line)
			}
			likeTables[currentTable] = parts[1]
		case strings.HasPrefix(line, "PRIMARY KEY"):
			continue
		default:
			column := strings.Fields(strings.TrimSuffix(line, ","))[0]
			currentColumns = append(currentColumns, column)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scan migration: %v", err)
	}

	for table, source := range likeTables {
		columns, ok := resolveLikeColumns(table, source, tables, likeTables, map[string]bool{})
		if !ok {
			t.Fatalf("failed to resolve LIKE columns for %s from %s", table, source)
		}
		tables[table] = columns
	}

	return tables
}

func resolveLikeColumns(table string, source string, tables map[string][]string, likeTables map[string]string, seen map[string]bool) ([]string, bool) {
	if columns, ok := tables[source]; ok && len(columns) > 0 {
		return append([]string(nil), columns...), true
	}

	if seen[table] {
		return nil, false
	}
	seen[table] = true

	nextSource, ok := likeTables[source]
	if !ok {
		return nil, false
	}

	columns, ok := resolveLikeColumns(source, nextSource, tables, likeTables, seen)
	if !ok {
		return nil, false
	}

	tables[source] = append([]string(nil), columns...)
	return append([]string(nil), columns...), true
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatalf("repo root not found from %s", wd)
	return ""
}
