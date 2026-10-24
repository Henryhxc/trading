package mysqlrepo

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/market"
	"ux-securities/backend/internal/migrations"
)

func TestNativeHistoryChunkCompletesWithinDeliveryDeadline(t *testing.T) {
	db := openIngestionTestDB(t)
	resetIngestionTestSchema(t, db)
	if err := migrations.Apply(context.Background(), db, filepath.Join("..", "..", "..", "..", "..", "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	collected := time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC)
	base := time.Date(2026, 10, 23, 1, 30, 0, 0, time.UTC)
	batch := domain.IngestionBatch{Version: 1}
	for i := 0; i < 400; i++ {
		bar := ingestionRepositoryBatch(collected, "380.2000").Candles[0]
		stamp := base.Add(time.Duration(i) * time.Minute)
		bar.Interval, bar.Timestamp = "1m", &stamp
		batch.Candles = append(batch.Candles, bar)
	}
	// The deployed sender gives each signed webhook 15 seconds. Leave headroom
	// for HTTP validation and commit instead of accepting incomplete backfills.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()
	repo := NewMarketRepository(db)
	result, err := repo.UpsertMarketBatch(ctx, batch)
	if err != nil {
		t.Fatalf("400-bar transaction exceeded delivery budget after %s: %v", time.Since(start), err)
	}
	if len(result.Events) != 1 || result.Events[0].Interval != "1m" {
		t.Fatalf("expected one committed period event: %+v", result.Events)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM stock_intraday_quotes").Scan(&count); err != nil || count != 400 {
		t.Fatalf("partial import: count=%d err=%v", count, err)
	}
	result, err = repo.UpsertMarketBatch(ctx, batch)
	if err != nil || len(result.Events) != 0 {
		t.Fatalf("identical retry should be a no-op: %+v %v", result, err)
	}
	t.Logf("400-bar insert and idempotent replay completed in %s", time.Since(start))
}

func TestMixedHistoryGroupsAreAtomicAndEmitOncePerPeriod(t *testing.T) {
	db := openIngestionTestDB(t)
	resetIngestionTestSchema(t, db)
	ctx := context.Background()
	if err := migrations.Apply(ctx, db, filepath.Join("..", "..", "..", "..", "..", "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("DELETE FROM stock_daily_quotes; DELETE FROM stocks"); err != nil {
		t.Fatal(err)
	}
	collected := time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC)
	batch := domain.IngestionBatch{Version: 1}
	for _, symbol := range []string{"00700", "00005"} {
		for _, interval := range []string{"1d", "1m", "5m"} {
			for i := 0; i < 2; i++ {
				bar := ingestionRepositoryBatch(collected.Add(time.Duration(i)*time.Second), "380.2000").Candles[0]
				bar.Symbol, bar.Interval = symbol, interval
				if interval != "1d" {
					stamp := time.Date(2026, 10, 23, 1, 30+i*5, 0, 0, time.UTC)
					bar.Timestamp = &stamp
				}
				batch.Candles = append(batch.Candles, bar)
			}
		}
	}
	// The final group fails after earlier daily and minute groups have written.
	invalid := batch
	invalid.Candles = append([]domain.CandleUpdate(nil), batch.Candles...)
	invalid.Candles[len(invalid.Candles)-1].TradeDate = "invalid-date"
	repo := NewMarketRepository(db)
	if _, err := repo.UpsertMarketBatch(ctx, invalid); err == nil {
		t.Fatal("expected final group to fail")
	}
	for _, table := range []string{"stocks", "stock_daily_quotes", "stock_intraday_quotes"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("mixed transaction leaked %s rows=%d err=%v", table, count, err)
		}
	}
	result, err := repo.UpsertMarketBatch(ctx, batch)
	if err != nil || len(result.Events) != 6 {
		t.Fatalf("mixed groups events=%+v err=%v", result.Events, err)
	}
	seen := map[string]bool{}
	for _, event := range result.Events {
		key := event.Symbol + ":" + event.Interval
		if seen[key] || !event.AsOf.Equal(collected.Add(time.Second)) {
			t.Fatalf("duplicate or old event: %+v", event)
		}
		seen[key] = true
	}
	result, err = repo.UpsertMarketBatch(ctx, batch)
	if err != nil || len(result.Events) != 0 {
		t.Fatalf("mixed replay should emit no events: %+v %v", result, err)
	}
}
