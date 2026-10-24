package market

import (
	"encoding/json"
	"testing"

	domain "ux-securities/backend/internal/domain/market"
)

func minuteUpdate(t *testing.T, interval, timestamp, day string) domain.CandleUpdate {
	t.Helper()
	raw := `{"symbol":"00700","trade_date":"` + day + `","interval":"` + interval + `","timestamp":"` + timestamp + `","open":"10","high":"12","low":"9","close":"11","volume":"100","turnover":"1100","source":"akshare","data_status":"unofficial","collected_at":"2026-10-24T00:00:00Z","adjustment":"none"}`
	var candle domain.CandleUpdate
	// Omit absent optional timestamps, matching the actual webhook contract.
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatal(err)
	}
	if timestamp == "" {
		delete(payload, "timestamp")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &candle); err != nil {
		t.Fatal(err)
	}
	return candle
}

func TestMinuteIngestionSeparatesNativePeriodsAndTimes(t *testing.T) {
	batch := domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{
		minuteUpdate(t, "1m", "2026-10-23T01:31:00Z", "2026-10-23"),
		minuteUpdate(t, "1m", "2026-10-23T01:32:00Z", "2026-10-23"),
		minuteUpdate(t, "5m", "2026-10-23T01:35:00Z", "2026-10-23"),
	}}
	if err := validateIngestionBatch(batch); err != nil {
		t.Fatalf("valid native minutes rejected: %v", err)
	}
	batch.Candles = append(batch.Candles, batch.Candles[0])
	if err := validateIngestionBatch(batch); err == nil {
		t.Fatal("duplicate minute accepted")
	}
}

func TestMinuteIngestionRejectsInvalidPeriodTimeAndCalendarDate(t *testing.T) {
	for _, test := range []struct{ period, stamp, day string }{
		{"2m", "2026-10-23T01:31:00Z", "2026-10-23"},
		{"1m", "2026-10-23T01:31:00Z", "2026-10-22"},
		{"1m", "", "2026-10-23"},
		{"1d", "", "2026-02-31"},
	} {
		t.Run(test.period+test.stamp+test.day, func(t *testing.T) {
			if err := validateIngestionBatch(domain.IngestionBatch{Version: 1, Candles: []domain.CandleUpdate{minuteUpdate(t, test.period, test.stamp, test.day)}}); err == nil {
				t.Fatal("invalid candle accepted")
			}
		})
	}
}
