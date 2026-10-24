package httpapi

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/market"
	service "ux-securities/backend/internal/service/market"
)

// These tests catch ignored interval parameters, quote-derived provenance,
// incorrect OHLC aggregation, and replacement of missing minutes by daily bars.
type intervalProvider struct {
	nativeCalls int
	requested   string
	from        time.Time
}

func (*intervalProvider) Search(context.Context, string) ([]domain.Symbol, error) { return nil, nil }
func (*intervalProvider) Quote(context.Context, string) (domain.Quote, error) {
	return domain.Quote{Source: "different-quote-source"}, nil
}
func (p *intervalProvider) Candles(_ context.Context, symbol string, from, _ time.Time, _ int) ([]domain.Candle, error) {
	if symbol == "12345" {
		return nil, domain.ErrNotFound
	}
	p.from = from
	var bars []domain.Candle
	_ = json.Unmarshal([]byte(`[
 {"trade_date":"2026-10-01T00:00:00Z","open":10,"high":15,"low":9,"close":12,"volume":100,"turnover":1200,"source":"akshare","data_status":"unofficial","collected_at":"2026-10-24T00:00:00Z"},
 {"trade_date":"2026-10-02T00:00:00Z","open":12,"high":17,"low":8,"close":16,"volume":200,"turnover":3200,"source":"akshare","data_status":"unofficial","collected_at":"2026-10-24T00:00:00Z"},
 {"trade_date":"2026-10-05T00:00:00Z","open":16,"high":19,"low":14,"close":18,"volume":50,"turnover":900,"source":"akshare","data_status":"unofficial","collected_at":"2026-10-24T00:00:00Z"}]`), &bars)
	return bars, nil
}
func (p *intervalProvider) NativeCandles(_ context.Context, symbol string, interval string, _, _ time.Time, _ int) ([]domain.Candle, error) {
	if symbol == "12345" {
		return nil, domain.ErrNotFound
	}
	p.nativeCalls++
	p.requested = interval
	return []domain.Candle{}, nil
}
func intervalRequest(t *testing.T, query string) (int, map[string]any, *intervalProvider) {
	t.Helper()
	p := &intervalProvider{}
	clock := func() time.Time { return time.Date(2026, 10, 24, 0, 1, 0, 0, time.UTC) }
	h := NewWithMarket(&testAuthService{}, false, service.NewServiceWithClock(p, time.Minute*20, clock), nil)
	r := httptest.NewRecorder()
	h.ServeHTTP(r, authenticatedRequest("GET", "/api/v1/markets/candles/00700?"+query, nil))
	var body map[string]any
	if err := json.Unmarshal(r.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return r.Code, body, p
}
func TestCandleIntervalRejectsUnknownDuplicateAndOversizedRange(t *testing.T) {
	for _, q := range []string{"interval=7m", "interval=1m&interval=1d", "interval=1d&from=2024-01-01&to=2026-10-24", "interval=1d&limit=5001"} {
		t.Run(q, func(t *testing.T) {
			code, _, _ := intervalRequest(t, q)
			if code != 400 {
				t.Fatalf("status=%d want 400", code)
			}
		})
	}
}
func TestCalendarCandlesAggregateOHLCAndVolume(t *testing.T) {
	for _, tc := range []struct {
		interval      string
		count         int
		firstDate     string
		close, volume float64
	}{
		{"1w", 2, "2026-09-28", 16, 300}, {"1M", 1, "2026-10-01", 18, 350}, {"1y", 1, "2026-01-01", 18, 350},
	} {
		t.Run(tc.interval, func(t *testing.T) {
			code, body, _ := intervalRequest(t, "interval="+tc.interval+"&from=2026-10-01&to=2026-10-24")
			if code != 200 || body["interval"] != tc.interval {
				t.Fatalf("status/body=%d/%v", code, body)
			}
			items := body["items"].([]any)
			if len(items) != tc.count {
				t.Fatalf("bars=%d want %d", len(items), tc.count)
			}
			first := items[0].(map[string]any)
			if first["open"] != float64(10) || first["high"] != float64(map[bool]int{true: 17, false: 19}[tc.interval == "1w"]) || first["low"] != float64(8) || first["close"] != tc.close || first["volume"] != tc.volume {
				t.Fatalf("bad OHLCV: %v", first)
			}
			if first["trade_date"].(string)[:10] != tc.firstDate {
				t.Fatalf("bucket=%v", first)
			}
			if body["source"] != "akshare" || body["data_status"] != "unofficial" {
				t.Fatalf("bad provenance %v", body)
			}
		})
	}
}
func TestMissingMinuteHistoryDoesNotSubstituteDailyCandles(t *testing.T) {
	for _, interval := range []string{"1m", "5m", "15m", "30m", "60m"} {
		t.Run(interval, func(t *testing.T) {
			code, body, p := intervalRequest(t, "interval="+interval+"&from=2026-04-24&to=2026-10-24")
			if code != 200 {
				t.Fatalf("status=%d: %v", code, body)
			}
			if len(body["items"].([]any)) != 0 || body["data_status"] != "unavailable" || p.nativeCalls != 1 || p.requested != interval {
				t.Fatalf("minute substitution: %v calls=%d", body, p.nativeCalls)
			}
		})
	}
}

func TestDefaultCandleWindowClampsMonthEnd(t *testing.T) {
	code, _, p := intervalRequest(t, "interval=1d&to=2026-08-31")
	if code != 200 {
		t.Fatalf("status=%d", code)
	}
	if got := p.from.Format("2006-01-02"); got != "2026-02-28" {
		t.Fatalf("default window start=%s want 2026-02-28", got)
	}
}

func TestExplicitCandleIntervalsPreserveNotFound(t *testing.T) {
	h := NewWithMarket(&testAuthService{}, false, service.NewServiceWithClock(&intervalProvider{}, time.Minute*20, time.Now), nil)
	for _, interval := range []string{"1d", "1w", "1M", "1y", "1m", "5m"} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, authenticatedRequest("GET", "/api/v1/markets/candles/12345?interval="+interval, nil))
		if r.Code != 404 {
			t.Fatalf("interval %s status=%d want 404", interval, r.Code)
		}
	}
}
