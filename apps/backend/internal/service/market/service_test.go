package market

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/market"
)

type providerStub struct {
	searchQuery string
	quoteSymbol string
	searchItems []domain.Symbol
	quote       domain.Quote
	candles     []domain.Candle
	err         error
}

func (p *providerStub) Search(_ context.Context, query string) ([]domain.Symbol, error) {
	p.searchQuery = query
	return p.searchItems, p.err
}
func (p *providerStub) Quote(_ context.Context, symbol string) (domain.Quote, error) {
	p.quoteSymbol = symbol
	return p.quote, p.err
}
func (p *providerStub) Candles(_ context.Context, _ string, _, _ time.Time, _ int) ([]domain.Candle, error) {
	return p.candles, p.err
}

func TestSearchTrimsQueryBeforeProviderCall(t *testing.T) {
	provider := &providerStub{searchItems: []domain.Symbol{{Symbol: "00700"}}}
	service := NewService(provider)
	items, err := service.Search(context.Background(), "  Tencent  ")
	if err != nil || len(items) != 1 || provider.searchQuery != "Tencent" {
		t.Fatalf("Search() items=%v err=%v query=%q", items, err, provider.searchQuery)
	}
}

func TestQuotePropagatesProviderFailure(t *testing.T) {
	want := errors.New("provider timeout")
	provider := &providerStub{err: want}
	_, err := NewService(provider).Quote(context.Background(), "00700")
	if !errors.Is(err, want) {
		t.Fatalf("Quote() error=%v, want %v", err, want)
	}
}

func TestQuoteDowngradesExpiredUnofficialDataToStale(t *testing.T) {
	price := 380.2
	provider := &providerStub{quote: domain.Quote{
		Symbol:     "00700",
		LastPrice:  &price,
		Source:     "akshare",
		DataStatus: domain.DataStatusUnofficial,
		AsOf:       time.Now().UTC().Add(-time.Minute),
	}}

	quote, err := NewService(provider).Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("Quote() error = %v, want nil", err)
	}
	if quote.DataStatus != domain.DataStatusStale {
		t.Fatalf("expired quote status = %q, want stale", quote.DataStatus)
	}
}

func TestCandlesDelegatesDailySeries(t *testing.T) {
	want := []domain.Candle{{Close: 380.2}}
	provider := &providerStub{candles: want}
	got, err := NewService(provider).Candles(context.Background(), "00700", time.Time{}, time.Time{}, 30)
	if err != nil || len(got) != 1 || got[0].Close != want[0].Close {
		t.Fatalf("Candles() got=%v err=%v", got, err)
	}
}

func TestDirectoryQuoteFreshnessMatchesDetail(t *testing.T) {
	now := time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC)
	price := 11.0
	provider := &providerStub{searchItems: []domain.Symbol{{Symbol: "00700", Quote: &domain.Quote{Symbol: "00700", LastPrice: &price, Source: "akshare", DataStatus: domain.DataStatusDelayed, AsOf: now.Add(-time.Hour)}}}}
	items, err := NewServiceWithClock(provider, time.Minute, func() time.Time { return now }).Search(context.Background(), "")
	if err != nil || len(items) != 1 {
		t.Fatalf("directory=%v err=%v", items, err)
	}
	if items[0].Quote.DataStatus != domain.DataStatusStale {
		t.Fatalf("directory status=%s, want stale", items[0].Quote.DataStatus)
	}
}
