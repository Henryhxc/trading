package watchlist

import (
	"context"
	"testing"

	auth "ux-securities/backend/internal/domain/auth"
	market "ux-securities/backend/internal/domain/market"
	domain "ux-securities/backend/internal/domain/watchlist"
)

type watchlistRepoStub struct {
	listUser, addUser, deleteUser, reorderUser int64
	addSymbol, deleteSymbol                    string
	items                                      []domain.Item
}

func (r *watchlistRepoStub) ListWatchlist(_ context.Context, userID int64) ([]domain.Item, error) {
	r.listUser = userID
	return append([]domain.Item(nil), r.items...), nil
}
func (r *watchlistRepoStub) AddWatchlist(_ context.Context, userID int64, symbol string) (domain.Item, bool, error) {
	r.addUser, r.addSymbol = userID, symbol
	return domain.Item{UserID: userID, Symbol: symbol}, true, nil
}
func (r *watchlistRepoStub) DeleteWatchlist(_ context.Context, userID int64, symbol string) error {
	r.deleteUser, r.deleteSymbol = userID, symbol
	return nil
}
func (r *watchlistRepoStub) ReorderWatchlist(_ context.Context, userID int64, _ []string) error {
	r.reorderUser = userID
	return nil
}

type quoteStub struct{}

func (quoteStub) Quote(_ context.Context, symbol string) (market.Quote, error) {
	price := 380.2
	return market.Quote{Symbol: symbol, LastPrice: &price, DataStatus: market.DataStatusSimulated}, nil
}

func TestAddNormalizesSymbolAndUsesActor(t *testing.T) {
	repo := &watchlistRepoStub{}
	actor := auth.Actor{UserID: 42}
	item, _, err := NewService(repo, quoteStub{}).Add(context.Background(), actor, "700")
	if err != nil || item.Symbol != "00700" || repo.addUser != 42 || repo.addSymbol != "00700" {
		t.Fatalf("Add() item=%+v err=%v repo=%+v", item, err, repo)
	}
}

func TestListLoadsQuotesForActorOwnedItems(t *testing.T) {
	repo := &watchlistRepoStub{items: []domain.Item{{UserID: 9, Symbol: "00700"}}}
	items, err := NewService(repo, quoteStub{}).List(context.Background(), auth.Actor{UserID: 9})
	if err != nil || repo.listUser != 9 || len(items) != 1 || items[0].Quote.LastPrice == nil {
		t.Fatalf("List() items=%+v err=%v repo=%+v", items, err, repo)
	}
}

func TestDeleteUsesActorAndNormalizedSymbol(t *testing.T) {
	repo := &watchlistRepoStub{}
	err := NewService(repo, quoteStub{}).Delete(context.Background(), auth.Actor{UserID: 7}, "5")
	if err != nil || repo.deleteUser != 7 || repo.deleteSymbol != "00005" {
		t.Fatalf("Delete() err=%v repo=%+v", err, repo)
	}
}
