package marketdata

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	market "ux-securities/backend/internal/domain/market"
)

// ErrUnavailable remains available to integration callers while the canonical
// error identity lives in the market domain for transport-layer mapping.
var ErrUnavailable = market.ErrUnavailable

// ErrNotFound remains available to integration callers while the canonical
// error identity lives in the market domain.
var ErrNotFound = market.ErrNotFound

type Provider struct {
	authorized market.MarketDataProvider
	unofficial market.MarketDataProvider
	mock       market.MarketDataProvider
}

func New(authorized, unofficial, mock market.MarketDataProvider) *Provider {
	return &Provider{authorized: authorized, unofficial: unofficial, mock: mock}
}

func (p *Provider) Search(ctx context.Context, query string) ([]market.Symbol, error) {
	for _, candidate := range []market.MarketDataProvider{p.authorized, p.unofficial, p.mock} {
		if candidate == nil {
			continue
		}
		items, err := candidate.Search(ctx, query)
		if err == nil {
			return items, nil
		}
	}
	return nil, ErrUnavailable
}

func (p *Provider) Quote(ctx context.Context, symbol string) (market.Quote, error) {
	var sawUnavailable, sawNotFound bool
	for index, candidate := range []market.MarketDataProvider{p.authorized, p.unofficial, p.mock} {
		if candidate == nil {
			continue
		}
		quote, err := candidate.Quote(ctx, symbol)
		if err == nil {
			if quote.DataStatus == market.DataStatusUnavailable && quote.LastPrice != nil {
				sawUnavailable = true
				continue
			}
			// Only an explicitly authorized provider may ever return live data.
			if index == 1 {
				quote.DataStatus = safeUnofficialStatus(quote.DataStatus)
			} else if index == 2 && quote.DataStatus == market.DataStatusLive {
				quote.DataStatus = market.DataStatusStale
			}
			return quote, nil
		}
		if errors.Is(err, market.ErrNotFound) {
			sawNotFound = true
		} else {
			sawUnavailable = true
		}
	}
	if sawNotFound && !sawUnavailable {
		return market.Quote{Symbol: normalize(symbol)}, market.ErrNotFound
	}
	return market.Quote{Symbol: normalize(symbol), DataStatus: market.DataStatusUnavailable, Source: "fallback", AsOf: time.Now().UTC()}, ErrUnavailable
}

func safeUnofficialStatus(status market.DataStatus) market.DataStatus {
	switch status {
	case market.DataStatusUnofficial, market.DataStatusDelayed, market.DataStatusStale, market.DataStatusUnavailable:
		return status
	default:
		return market.DataStatusStale
	}
}

func (p *Provider) Candles(ctx context.Context, symbol string, from, to time.Time, limit int) ([]market.Candle, error) {
	var sawUnavailable, sawNotFound bool
	for _, candidate := range []market.MarketDataProvider{p.authorized, p.unofficial, p.mock} {
		if candidate == nil {
			continue
		}
		items, err := candidate.Candles(ctx, symbol, from, to, limit)
		if err == nil {
			return items, nil
		}
		if errors.Is(err, market.ErrNotFound) {
			sawNotFound = true
		} else {
			sawUnavailable = true
		}
	}
	if sawNotFound && !sawUnavailable {
		return nil, market.ErrNotFound
	}
	return nil, ErrUnavailable
}

type MockProvider struct{ symbols map[string]market.Symbol }

func (p *Provider) NativeCandles(ctx context.Context, symbol, interval string, from, to time.Time, limit int) ([]market.Candle, error) {
	for _, candidate := range []market.MarketDataProvider{p.authorized, p.unofficial} {
		if native, ok := candidate.(market.NativeCandleProvider); ok {
			return native.NativeCandles(ctx, symbol, interval, from, to, limit)
		}
	}
	// A missing native adapter is an empty series only for a known symbol.
	// Check identity without manufacturing mock minute candles.
	if _, err := p.Quote(ctx, symbol); err != nil {
		return nil, err
	}
	return []market.Candle{}, nil
}

func NewMockProvider() *MockProvider {
	return &MockProvider{symbols: map[string]market.Symbol{
		"00700": {Symbol: "00700", DisplayName: "Tencent Holdings"},
		"00005": {Symbol: "00005", DisplayName: "HSBC Holdings"},
		"09988": {Symbol: "09988", DisplayName: "Alibaba Group"},
	}}
}

func (m *MockProvider) Search(_ context.Context, query string) ([]market.Symbol, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	items := make([]market.Symbol, 0)
	for _, item := range m.symbols {
		if strings.Contains(item.Symbol, query) || strings.Contains(strings.ToLower(item.DisplayName), query) {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Symbol < items[j].Symbol })
	return items, nil
}

func (m *MockProvider) Quote(_ context.Context, symbol string) (market.Quote, error) {
	symbol = normalize(symbol)
	item, ok := m.symbols[symbol]
	if !ok {
		return market.Quote{}, market.ErrNotFound
	}
	price := map[string]float64{"00700": 380.20, "00005": 63.45, "09988": 88.10}[symbol]
	change := map[string]float64{"00700": 2.10, "00005": -0.35, "09988": 1.20}[symbol]
	percent := change / (price - change) * 100
	return market.Quote{Symbol: item.Symbol, DisplayName: item.DisplayName, LastPrice: &price, Bid: ptr(price - 0.2), Ask: ptr(price + 0.2), Change: &change, ChangePercent: &percent, Volume: 1200000, Turnover: price * 1200000, AsOf: time.Now().UTC(), Source: "mock", DataStatus: market.DataStatusSimulated}, nil
}

func (m *MockProvider) Candles(_ context.Context, symbol string, from, to time.Time, limit int) ([]market.Candle, error) {
	if _, ok := m.symbols[normalize(symbol)]; !ok {
		return nil, market.ErrNotFound
	}
	if limit <= 0 || limit > 366 {
		limit = 30
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}
	items := make([]market.Candle, 0, limit)
	for i := limit - 1; i >= 0; i-- {
		day := to.AddDate(0, 0, -i)
		if !from.IsZero() && day.Before(from) {
			continue
		}
		base := 370.0 + float64(limit-i)
		items = append(items, market.Candle{TradeDate: day, Open: base, High: base + 2, Low: base - 2, Close: base + 1, Volume: 500000 + int64(i*1000), Turnover: ptr(base * 500000), Source: "mock", DataStatus: market.DataStatusSimulated, CollectedAt: time.Now().UTC()})
	}
	return items, nil
}

func ptr(value float64) *float64 { return &value }
func normalize(value string) string {
	value = strings.TrimSpace(value)
	for len(value) < 5 {
		value = "0" + value
	}
	return value
}
