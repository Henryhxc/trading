package watchlist

import "ux-securities/backend/internal/domain/market"

type Item struct {
	ID        int64        `json:"id"`
	UserID    int64        `json:"-"`
	Symbol    string       `json:"symbol"`
	Name      string       `json:"display_name"`
	SortOrder int          `json:"sort_order"`
	Quote     market.Quote `json:"quote"`
}
