package market

import (
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/market"
)

func TestAggregateMissingTurnoverRemainsNull(t *testing.T) {
	amount := 100.0
	items := []domain.Candle{
		{TradeDate: time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC), Turnover: &amount},
		{TradeDate: time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)},
	}
	for _, period := range []string{"1w", "1M", "1y"} {
		got := aggregateCandles(items, period)
		if len(got) != 1 || got[0].Turnover != nil {
			t.Fatalf("%s must retain missing amount: %+v", period, got)
		}
	}
	items[1].Turnover = &amount
	got := aggregateCandles(items, "1M")
	if got[0].Turnover == nil || *got[0].Turnover != 200 || amount != 100 {
		t.Fatalf("sum mutated inputs or lost amount: %+v original=%v", got, amount)
	}
}
