package mysqlrepo

import (
	"context"

	domain "ux-securities/backend/internal/domain/watchlist"
)

func (r *MarketRepository) ListWatchlist(ctx context.Context, userID int64) ([]domain.Item, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT w.id, w.user_id, s.symbol, s.name, w.sort_order FROM watchlists w JOIN stocks s ON s.id = w.stock_id WHERE w.user_id = ? ORDER BY w.sort_order, w.id", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Item{}
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.UserID, &item.Symbol, &item.Name, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MarketRepository) AddWatchlist(ctx context.Context, userID int64, symbol string) (domain.Item, bool, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, "SELECT id FROM stocks WHERE symbol = ? AND status = 'active'", symbol).Scan(&id)
	if err != nil {
		return domain.Item{}, false, err
	}
	result, err := r.db.ExecContext(ctx, "INSERT INTO watchlists (user_id, stock_id, sort_order, created_at) VALUES (?, ?, COALESCE((SELECT MAX(sort_order)+1 FROM watchlists w2 WHERE w2.user_id = ?), 0), UTC_TIMESTAMP(3)) ON DUPLICATE KEY UPDATE sort_order = sort_order", userID, id, userID)
	if err != nil {
		return domain.Item{}, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.Item{}, false, err
	}
	var item domain.Item
	err = r.db.QueryRowContext(ctx, "SELECT w.id, w.user_id, s.symbol, s.name, w.sort_order FROM watchlists w JOIN stocks s ON s.id = w.stock_id WHERE w.user_id = ? AND s.symbol = ?", userID, symbol).Scan(&item.ID, &item.UserID, &item.Symbol, &item.Name, &item.SortOrder)
	return item, affected > 0, err
}

func (r *MarketRepository) DeleteWatchlist(ctx context.Context, userID int64, symbol string) error {
	_, err := r.db.ExecContext(ctx, "DELETE w FROM watchlists w JOIN stocks s ON s.id = w.stock_id WHERE w.user_id = ? AND s.symbol = ?", userID, symbol)
	return err
}

func (r *MarketRepository) ReorderWatchlist(ctx context.Context, userID int64, symbols []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, symbol := range symbols {
		if _, err := tx.ExecContext(ctx, "UPDATE watchlists w JOIN stocks s ON s.id = w.stock_id SET w.sort_order = ? WHERE w.user_id = ? AND s.symbol = ?", i, userID, symbol); err != nil {
			return err
		}
	}
	return tx.Commit()
}
