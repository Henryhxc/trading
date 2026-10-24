package market

import (
	"context"
	"strings"
	"time"

	domain "ux-securities/backend/internal/domain/market"
)

const defaultStaleAfter = 30 * time.Second

type Service struct {
	provider   domain.MarketDataProvider
	staleAfter time.Duration
	now        func() time.Time
}

func NewService(provider domain.MarketDataProvider) *Service {
	return NewServiceWithClock(provider, defaultStaleAfter, time.Now)
}

func NewServiceWithClock(provider domain.MarketDataProvider, staleAfter time.Duration, now func() time.Time) *Service {
	return &Service{provider: provider, staleAfter: staleAfter, now: now}
}
func (s *Service) Search(ctx context.Context, query string) ([]domain.Symbol, error) {
	items, err := s.provider.Search(ctx, strings.TrimSpace(query))
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Quote != nil {
			quote := s.withFreshness(*items[i].Quote)
			items[i].Quote = &quote
		}
	}
	return items, nil
}
func (s *Service) Quote(ctx context.Context, symbol string) (domain.Quote, error) {
	quote, err := s.provider.Quote(ctx, symbol)
	if err != nil {
		return quote, err
	}
	return s.withFreshness(quote), nil
}

func (s *Service) withFreshness(quote domain.Quote) domain.Quote {
	if isUnofficialStatus(quote.DataStatus) && !quote.AsOf.IsZero() && s.now().Sub(quote.AsOf) > s.staleAfter {
		quote.DataStatus = domain.DataStatusStale
	}
	return quote
}
func (s *Service) Candles(ctx context.Context, symbol string, from, to time.Time, limit int) ([]domain.Candle, error) {
	return s.provider.Candles(ctx, symbol, from, to, limit)
}

func isUnofficialStatus(status domain.DataStatus) bool {
	return status == domain.DataStatusUnofficial || status == domain.DataStatusDelayed
}
