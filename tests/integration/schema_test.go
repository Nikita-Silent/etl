//go:build integration
// +build integration

package integration

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/models"
	"github.com/user/go-frontol-loader/tests/integration/framework"
)

func TestDatabaseSchemaMatchesMigrations(t *testing.T) {
	env := framework.SetupTestEnvironment(t)
	ctx := env.GetContext()

	tables := make([]string, 0, len(models.TxSchemas))
	for table := range models.TxSchemas {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	missing := make([]string, 0)
	for _, table := range tables {
		ok, err := hasConflictConstraint(ctx, env.Postgres.Pool, table)
		if err != nil {
			t.Fatalf("Failed to check constraints for %s: %v", table, err)
		}
		if !ok {
			missing = append(missing, table)
		}
	}

	if len(missing) > 0 {
		t.Fatalf("Missing UNIQUE/PK constraint on (transaction_id_unique, source_folder): %v", missing)
	}
}

func hasConflictConstraint(ctx context.Context, pool *db.Pool, table string) (bool, error) {
	query := `
		SELECT tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = 'public'
		  AND tc.table_name = $1
		  AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
		  AND kcu.column_name IN ('transaction_id_unique', 'source_folder')
		GROUP BY tc.constraint_name
		HAVING COUNT(DISTINCT kcu.column_name) = 2
	`

	row := pool.QueryRow(ctx, query, table)
	var constraintName string
	if err := row.Scan(&constraintName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return constraintName != "", nil
}
