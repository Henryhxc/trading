package migrations

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestRunnerReappliesMigrationAfterRollback(t *testing.T) {
	dsn := os.Getenv("PHASE1_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set PHASE1_MIGRATION_TEST_DSN to a disposable ux_phase1_test_* schema")
	}
	parts := strings.Split(strings.SplitN(dsn, "/", 2)[1], "?")
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "ux_phase1_test_") {
		t.Fatal("refusing to run migration rollback outside a ux_phase1_test_* schema")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	for _, table := range []string{"accounts", "sessions", "users", "schema_migrations"} {
		if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			t.Fatalf("clear disposable test table %s: %v", table, err)
		}
	}

	root := filepath.Clean(filepath.Join("..", "..", "..", ".."))
	migrationDir := filepath.Join(root, "db", "migrations")
	if err := Apply(ctx, db, migrationDir); err != nil {
		t.Fatalf("initial Apply(): %v", err)
	}
	down, err := os.ReadFile(filepath.Join(migrationDir, "001_users_sessions_accounts.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range strings.Split(string(down), ";") {
		if statement = strings.TrimSpace(statement); statement != "" {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				t.Fatalf("rollback statement %q: %v", statement, err)
			}
		}
	}
	if err := Apply(ctx, db, migrationDir); err != nil {
		t.Fatalf("reapply after down migration: %v", err)
	}
	for _, table := range []string{"accounts", "sessions", "users"} {
		var exists int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists != 1 {
			t.Errorf("table %s exists after runner reapply = %d, want 1", table, exists)
		}
	}
}
