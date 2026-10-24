package watchlist

import (
	"context"
	"errors"
	"strings"
	"time"

	"ux-securities/backend/internal/domain/auth"
	market "ux-securities/backend/internal/domain/market"
	domain "ux-securities/backend/internal/domain/watchlist"
)

var ErrNotFound = errors.New("watchlist item not found")

type Repository interface {
	ListWatchlist(context.Context, int64) ([]domain.Item, error)
	AddWatchlist(context.Context, int64, string) (domain.Item, bool, error)
	DeleteWatchlist(context.Context, int64, string) error
	ReorderWatchlist(context.Context, int64, []string) error
}
type MarketService interface {
	Quote(context.Context, string) (market.Quote, error)
}
type Service struct {
	repo   Repository
	market MarketService
}

func NewService(repo Repository, markets MarketService) *Service {
	return &Service{repo: repo, market: markets}
}
func (s *Service) List(ctx context.Context, actor auth.Actor) ([]domain.Item, error) {
	items, err := s.repo.ListWatchlist(ctx, actor.UserID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		quote, quoteErr := s.market.Quote(ctx, items[i].Symbol)
		if quoteErr != nil {
			quote = market.Quote{Symbol: items[i].Symbol, DisplayName: items[i].Name, Source: "fallback", DataStatus: market.DataStatusUnavailable, AsOf: time.Now().UTC()}
		}
		items[i].Quote = quote
	}
	return items, nil
}
func (s *Service) Add(ctx context.Context, actor auth.Actor, symbol string) (domain.Item, bool, error) {
	return s.repo.AddWatchlist(ctx, actor.UserID, normalize(symbol))
}
func (s *Service) Delete(ctx context.Context, actor auth.Actor, symbol string) error {
	return s.repo.DeleteWatchlist(ctx, actor.UserID, normalize(symbol))
}
func (s *Service) Reorder(ctx context.Context, actor auth.Actor, symbols []string) error {
	for i := range symbols {
		symbols[i] = normalize(symbols[i])
	}
	return s.repo.ReorderWatchlist(ctx, actor.UserID, symbols)
}
func normalize(symbol string) string {
	symbol = strings.TrimSpace(symbol)
	for len(symbol) < 5 {
		symbol = "0" + symbol
	}
	return symbol
}
