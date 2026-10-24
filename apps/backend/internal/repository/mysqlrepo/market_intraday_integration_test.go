package mysqlrepo

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/market"
	"ux-securities/backend/internal/migrations"
)

func TestNativeMinuteStorageProvenanceUniquenessAndOlderOverwrite(t *testing.T) {
	db := openIngestionTestDB(t)
	resetIngestionTestSchema(t, db)
	ctx := context.Background()
	if err := migrations.Apply(ctx, db, filepath.Join("..", "..", "..", "..", "..", "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	// Simulate a crash after MySQL committed DDL but before the ledger write.
	if _, err := db.Exec("DELETE FROM schema_migrations WHERE version='004_intraday_candles.sql'"); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Apply(ctx, db, filepath.Join("..", "..", "..", "..", "..", "db", "migrations")); err != nil {
		t.Fatalf("migration recovery: %v", err)
	}
	repoMissing := NewMarketRepository(db)
	if _, err := repoMissing.NativeCandles(ctx, "12345", "1m", time.Now().AddDate(0, -6, 0), time.Now(), 5000); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown minute symbol: %v", err)
	}
	// Use a symbol without a quote to prove history is independently importable.
	now := time.Date(2026, 10, 24, 2, 0, 0, 0, time.UTC)
	stamp := time.Date(2026, 10, 23, 1, 35, 0, 0, time.UTC)
	bar := ingestionRepositoryBatch(now, "380.2000").Candles[0]
	bar.Symbol = "01810"
	bar.Interval = "5m"
	bar.Timestamp = &stamp
	repo := NewMarketRepository(db)
	batch := domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{bar}}
	for i := 0; i < 2; i++ {
		if _, err := repo.UpsertMarketBatch(ctx, batch); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM stock_intraday_quotes").Scan(&count); err != nil || count != 1 {
		t.Fatalf("unique count=%d err=%v", count, err)
	}
	older := bar
	older.CollectedAt = now.Add(-time.Hour)
	older.Close = "1.0000"
	result, err := repo.UpsertMarketBatch(ctx, domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{older}})
	if err != nil || len(result.Events) != 0 {
		t.Fatalf("old batch=%+v err=%v", result, err)
	}
	items, err := repo.NativeCandles(ctx, "01810", "5m", now.AddDate(0, 0, -2), now, 5000)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	if items[0].Source != "akshare" || items[0].DataStatus != domain.DataStatusUnofficial || items[0].Close != 380.2 || !items[0].Timestamp.Equal(stamp) || items[0].Volume != 1000000 {
		t.Fatalf("lost native candle: %+v", items[0])
	}
	wrongPeriod, err := repo.NativeCandles(ctx, "01810", "1m", now.AddDate(0, 0, -2), now, 5000)
	if err != nil || len(wrongPeriod) != 0 {
		t.Fatalf("leaked period: %v %v", wrongPeriod, err)
	}
	// Daily history also rejects out-of-order overwrites and retains provenance.
	daily := bar
	daily.Interval = ""
	daily.Timestamp = nil
	if _, err = repo.UpsertMarketBatch(ctx, domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{daily}}); err != nil {
		t.Fatal(err)
	}
	daily.CollectedAt = now.Add(-time.Minute)
	daily.Close = "1"
	if _, err = repo.UpsertMarketBatch(ctx, domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{daily}}); err != nil {
		t.Fatal(err)
	}
	days, err := repo.Candles(ctx, "01810", now.AddDate(0, 0, -2), now, 5000)
	if err != nil || len(days) != 1 || days[0].Close != 380.2 || days[0].Source != "akshare" {
		t.Fatalf("daily=%+v err=%v", days, err)
	}
	// Public daily records have no amount; NULL must survive storage and reads.
	daily.CollectedAt = now.Add(time.Minute)
	daily.Source = "tencent"
	daily.Close = "380.2000"
	daily.Turnover = ""
	if _, err = repo.UpsertMarketBatch(ctx, domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{daily}}); err != nil {
		t.Fatal(err)
	}
	days, err = repo.Candles(ctx, "01810", now.AddDate(0, 0, -2), now, 5000)
	if err != nil || len(days) != 1 || days[0].Turnover != nil || days[0].Source != "tencent" {
		t.Fatalf("missing turnover/provenance lost: %+v %v", days, err)
	}
	// A later failing minute write must roll back an earlier quote in the batch.
	bar.Source = "yahoo"
	bar.Turnover = ""
	bar.CollectedAt = now.Add(time.Minute)
	if _, err = repo.UpsertMarketBatch(ctx, domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{bar}}); err != nil {
		t.Fatal(err)
	}
	items, err = repo.NativeCandles(ctx, "01810", "5m", now.AddDate(0, 0, -2), now, 5000)
	if err != nil || len(items) != 1 || items[0].Turnover != nil || items[0].Source != "yahoo" {
		t.Fatalf("native missing amount/provenance lost: %+v %v", items, err)
	}
	// A later failing minute write must roll back an earlier quote in the batch.
	broken := ingestionRepositoryBatch(now.Add(time.Minute), "400.0000")
	invalid := bar
	invalid.TradeDate = "not-a-date"
	broken.Candles = []domain.CandleUpdate{invalid}
	if _, err = repo.UpsertMarketBatch(ctx, broken); err == nil {
		t.Fatal("invalid minute should fail")
	}
	quote, err := repo.Quote(ctx, "00700")
	if err != nil || *quote.LastPrice == 400 {
		t.Fatalf("quote was not rolled back: %+v %v", quote, err)
	}
}
