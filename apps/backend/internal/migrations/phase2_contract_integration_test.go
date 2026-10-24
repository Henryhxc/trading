package migrations

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	canonicalPhase2Up    = "002_stocks_daily_quotes_watchlists.sql"
	canonicalPhase2Down  = "002_stocks_daily_quotes_watchlists.down.sql"
	provenancePhase2Up   = "002_webhook_market_provenance.sql"
	provenancePhase2Down = "002_webhook_market_provenance.down.sql"
)

func TestPhase2MigrationContractLifecycle(t *testing.T) {
	db, ctx := openPhase2MigrationTestDB(t)
	root := filepath.Clean(filepath.Join("..", "..", "..", ".."))
	migrationDir := filepath.Join(root, "db", "migrations")
	phase2Dir := t.TempDir()
	copyMigrationFile(t, migrationDir, phase2Dir, "001_users_sessions_accounts.sql")
	copyMigrationFile(t, migrationDir, phase2Dir, canonicalPhase2Up)
	copyMigrationFile(t, migrationDir, phase2Dir, provenancePhase2Up)
	t.Cleanup(func() { resetPhase2MigrationTestSchema(t, ctx, db) })

	t.Run("fresh apply, rollback, and reapply use the canonical ledger", func(t *testing.T) {
		resetPhase2MigrationTestSchema(t, ctx, db)
		if err := Apply(ctx, db, phase2Dir); err != nil {
			t.Fatalf("fresh Apply(): %v", err)
		}
		assertTables(t, ctx, db, []string{
			"accounts", "schema_migrations", "sessions", "stock_daily_quotes",
			"stocks", "users", "watchlists",
		})
		assertMigrationLedger(t, ctx, db, []string{
			"001_users_sessions_accounts.sql", canonicalPhase2Up, provenancePhase2Up,
		})

		executeSQLFile(t, ctx, db, filepath.Join(migrationDir, provenancePhase2Down))
		executeSQLFile(t, ctx, db, filepath.Join(migrationDir, canonicalPhase2Down))
		assertTables(t, ctx, db, []string{
			"accounts", "schema_migrations", "sessions", "users",
		})
		assertMigrationLedger(t, ctx, db, []string{"001_users_sessions_accounts.sql"})

		if err := Apply(ctx, db, phase2Dir); err != nil {
			t.Fatalf("reapply after canonical Phase 2 rollback: %v", err)
		}
		assertTables(t, ctx, db, []string{
			"accounts", "schema_migrations", "sessions", "stock_daily_quotes",
			"stocks", "users", "watchlists",
		})
		assertMigrationLedger(t, ctx, db, []string{
			"001_users_sessions_accounts.sql", canonicalPhase2Up, provenancePhase2Up,
		})
	})

	t.Run("legacy split ledger upgrades and leaves 003 available", func(t *testing.T) {
		resetPhase2MigrationTestSchema(t, ctx, db)
		if err := Apply(ctx, db, phase2Dir); err != nil {
			t.Fatalf("prepare canonical tables: %v", err)
		}
		if _, err := db.ExecContext(ctx,
			"DELETE FROM schema_migrations WHERE version = ?", canonicalPhase2Up,
		); err != nil {
			t.Fatal(err)
		}
		for _, legacyVersion := range []string{
			"002_stocks_daily_quotes.sql", "003_watchlists.sql",
		} {
			if _, err := db.ExecContext(ctx,
				"INSERT INTO schema_migrations (version) VALUES (?)", legacyVersion,
			); err != nil {
				t.Fatalf("seed legacy ledger %s: %v", legacyVersion, err)
			}
		}

		futureDir := t.TempDir()
		copyMigrationFile(t, migrationDir, futureDir, "001_users_sessions_accounts.sql")
		copyMigrationFile(t, migrationDir, futureDir, canonicalPhase2Up)
		copyMigrationFile(t, migrationDir, futureDir, provenancePhase2Up)
		phase3Probe := filepath.Join(futureDir, "003_orders_positions.sql")
		if err := os.WriteFile(phase3Probe, []byte(
			"CREATE TABLE migration_phase3_probe (id INT NOT NULL PRIMARY KEY);\n",
		), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := Apply(ctx, db, futureDir); err != nil {
			t.Fatalf("upgrade legacy ledger and apply future 003: %v", err)
		}
		assertMigrationLedger(t, ctx, db, []string{
			"001_users_sessions_accounts.sql",
			canonicalPhase2Up,
			provenancePhase2Up,
			"003_orders_positions.sql",
		})
		assertTableExists(t, ctx, db, "migration_phase3_probe", true)

		executeSQLFile(t, ctx, db, filepath.Join(migrationDir, provenancePhase2Down))
		executeSQLFile(t, ctx, db, filepath.Join(migrationDir, canonicalPhase2Down))
		assertTables(t, ctx, db, []string{
			"accounts", "migration_phase3_probe", "schema_migrations", "sessions", "users",
		})
		assertMigrationLedger(t, ctx, db, []string{
			"001_users_sessions_accounts.sql", "003_orders_positions.sql",
		})

		if err := Apply(ctx, db, futureDir); err != nil {
			t.Fatalf("reapply canonical Phase 2 after upgraded-ledger rollback: %v", err)
		}
		assertTables(t, ctx, db, []string{
			"accounts", "migration_phase3_probe", "schema_migrations", "sessions",
			"stock_daily_quotes", "stocks", "users", "watchlists",
		})
		assertMigrationLedger(t, ctx, db, []string{
			"001_users_sessions_accounts.sql",
			canonicalPhase2Up,
			provenancePhase2Up,
			"003_orders_positions.sql",
		})
	})
}

func openPhase2MigrationTestDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	dsn := os.Getenv("PHASE2_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set PHASE2_MIGRATION_TEST_DSN to a disposable ux_phase2_migration_test_* schema")
	}
	dsnParts := strings.SplitN(dsn, "/", 2)
	if len(dsnParts) != 2 {
		t.Fatal("PHASE2_MIGRATION_TEST_DSN has no database path")
	}
	databaseName := strings.Split(dsnParts[1], "?")[0]
	if !strings.HasPrefix(databaseName, "ux_phase2_migration_test_") {
		t.Fatal("refusing to run migration tests outside a ux_phase2_migration_test_* schema")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	return db, ctx
}

func resetPhase2MigrationTestSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{
		"migration_phase3_probe", "watchlists", "stock_daily_quotes", "stocks",
		"accounts", "sessions", "users", "schema_migrations",
	} {
		if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			t.Fatalf("drop disposable table %s: %v", table, err)
		}
	}
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatal(err)
	}
}

func executeSQLFile(t *testing.T, ctx context.Context, db *sql.DB, filename string) {
	t.Helper()
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range strings.Split(string(source), ";") {
		if statement = strings.TrimSpace(statement); statement != "" {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				t.Fatalf("execute %s statement %q: %v", filepath.Base(filename), statement, err)
			}
		}
	}
}

func copyMigrationFile(t *testing.T, fromDir, toDir, name string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(fromDir, name))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toDir, name), source, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertMigrationLedger(t *testing.T, ctx context.Context, db *sql.DB, want []string) {
	t.Helper()
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			t.Fatal(err)
		}
		got = append(got, version)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("schema_migrations = %v, want %v", got, want)
	}
	t.Logf("schema_migrations: %v", got)
}

func assertTables(t *testing.T, ctx context.Context, db *sql.DB, want []string) {
	t.Helper()
	rows, err := db.QueryContext(ctx,
		"SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_name",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		got = append(got, table)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tables = %v, want %v", got, want)
	}
	t.Logf("tables: %v", got)
}

func assertTableExists(t *testing.T, ctx context.Context, db *sql.DB, table string, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?",
		table,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count == 1; got != want {
		t.Fatalf("table %s exists = %t, want %t", table, got, want)
	}
}
