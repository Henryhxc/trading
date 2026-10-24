package mysqlrepo

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	domain "ux-securities/backend/internal/domain/market"
	"ux-securities/backend/internal/migrations"
)

func TestUpsertMarketBatchCommitsProvenanceAndRejectsOlderOverwrite(t *testing.T) {
	db := openIngestionTestDB(t)
	ctx := context.Background()
	resetIngestionTestSchema(t, db)
	if err := migrations.Apply(ctx, db, filepath.Join("..", "..", "..", "..", "..", "db", "migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if _, err := db.Exec("DELETE FROM stock_daily_quotes; DELETE FROM stocks"); err != nil {
		t.Fatalf("clear market fixtures: %v", err)
	}
	repository := NewMarketRepository(db)
	newer := time.Date(2026, 10, 24, 2, 0, 0, 0, time.UTC)
	batch := ingestionRepositoryBatch(newer, "380.2000")

	result, err := repository.UpsertMarketBatch(ctx, batch)
	if err != nil {
		t.Fatalf("UpsertMarketBatch() error = %v, want nil", err)
	}
	if len(result.Events) != 2 {
		t.Fatalf("events = %d, want quote and candles events", len(result.Events))
	}
	var price, source, status string
	var quoteAsOf time.Time
	if err := db.QueryRow("SELECT last_price, quote_source, data_status, quote_as_of FROM stocks WHERE symbol = '00700'").Scan(&price, &source, &status, &quoteAsOf); err != nil {
		t.Fatal(err)
	}
	if price != "380.2000" || source != "akshare" || status != "delayed" || !quoteAsOf.Equal(newer) {
		t.Fatalf("stored quote = %s/%s/%s/%s", price, source, status, quoteAsOf)
	}
	var dailySource, dailyStatus, adjustment string
	var collectedAt time.Time
	if err := db.QueryRow("SELECT source, data_status, collected_at, adjustment FROM stock_daily_quotes").Scan(&dailySource, &dailyStatus, &collectedAt, &adjustment); err != nil {
		t.Fatal(err)
	}
	if dailySource != "akshare" || dailyStatus != "unofficial" || adjustment != "none" || !collectedAt.Equal(newer) {
		t.Fatalf("daily provenance = %s/%s/%s/%s", dailySource, dailyStatus, adjustment, collectedAt)
	}

	older := ingestionRepositoryBatch(newer.Add(-time.Minute), "1.0000")
	older.Candles = nil
	result, err = repository.UpsertMarketBatch(ctx, older)
	if err != nil {
		t.Fatalf("older UpsertMarketBatch() error = %v", err)
	}
	if len(result.Events) != 0 {
		t.Fatalf("older events = %d, want 0", len(result.Events))
	}
	if err := db.QueryRow("SELECT last_price FROM stocks WHERE symbol = '00700'").Scan(&price); err != nil {
		t.Fatal(err)
	}
	if price != "380.2000" {
		t.Fatalf("price after older update = %s, want 380.2000", price)
	}
}

func TestUpsertMarketBatchRollsBackQuoteWhenCandleWriteFails(t *testing.T) {
	db := openIngestionTestDB(t)
	ctx := context.Background()
	resetIngestionTestSchema(t, db)
	if err := migrations.Apply(ctx, db, filepath.Join("..", "..", "..", "..", "..", "db", "migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if _, err := db.Exec("DELETE FROM stock_daily_quotes; DELETE FROM stocks"); err != nil {
		t.Fatalf("clear market fixtures: %v", err)
	}
	batch := ingestionRepositoryBatch(time.Date(2026, 10, 24, 2, 0, 0, 0, time.UTC), "380.2000")
	batch.Candles[0].TradeDate = "invalid-date"

	if _, err := NewMarketRepository(db).UpsertMarketBatch(ctx, batch); err == nil {
		t.Fatal("UpsertMarketBatch() error = nil, want candle failure")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM stocks WHERE symbol = '00700'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("stock rows after rollback = %d, want 0", count)
	}
}

func ingestionRepositoryBatch(collectedAt time.Time, price string) domain.IngestionBatch {
	return domain.IngestionBatch{Version: 1,
		Quotes:  []domain.QuoteUpdate{{Symbol: "00700", DisplayName: "Tencent Holdings", LastPrice: price, Change: "2.1000", ChangePercent: "0.5543", Volume: "1200000", Turnover: "456240000.0000", CollectedAt: collectedAt, Source: "akshare", DataStatus: domain.DataStatusDelayed}},
		Candles: []domain.CandleUpdate{{Symbol: "00700", TradeDate: "2026-10-23", Open: "376.0000", High: "382.0000", Low: "375.0000", Close: "380.2000", Volume: "1000000", Turnover: "379000000.0000", Source: "akshare", DataStatus: domain.DataStatusUnofficial, CollectedAt: collectedAt, Adjustment: "none"}},
	}
}

func openIngestionTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("INGESTION_TEST_DSN")
	if dsn == "" {
		t.Skip("set INGESTION_TEST_DSN to a disposable ux_phase2_ingestion_test_* schema")
	}
	databaseName := strings.Split(strings.SplitN(dsn, "/", 2)[1], "?")[0]
	if !strings.HasPrefix(databaseName, "ux_phase2_ingestion_test_") {
		t.Fatal("refusing to run outside a ux_phase2_ingestion_test_* schema")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func resetIngestionTestSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS positions, orders, watchlists, stock_intraday_quotes, stock_daily_quotes, stocks, accounts, sessions, users, schema_migrations; SET FOREIGN_KEY_CHECKS=1")
	if err != nil {
		t.Fatalf("reset disposable schema: %v", err)
	}
}
