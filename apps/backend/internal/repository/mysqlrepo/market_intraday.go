package mysqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	market "ux-securities/backend/internal/domain/market"
)

func (r *MarketRepository) NativeCandles(ctx context.Context, symbol, interval string, from, to time.Time, limit int) ([]market.Candle, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx, "SELECT 1 FROM stocks WHERE symbol = ?", symbol).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return nil, market.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	// Dates are exchange dates, while bar_time is always stored in UTC.
	rows, err := r.db.QueryContext(ctx, `SELECT q.trade_date, q.bar_time, q.open, q.high, q.low, q.close,
 q.volume, q.turnover, q.source, q.data_status, q.collected_at
 FROM stock_intraday_quotes q JOIN stocks s ON s.id=q.stock_id
 WHERE s.symbol=? AND q.interval_code=? AND q.trade_date>=? AND q.trade_date<=?
 ORDER BY q.bar_time DESC LIMIT ?`, symbol, interval, from.Format("2006-01-02"), to.Format("2006-01-02"), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []market.Candle{}
	for rows.Next() {
		var item market.Candle
		var stamp time.Time
		var volume sql.NullFloat64
		if err := rows.Scan(&item.TradeDate, &stamp, &item.Open, &item.High, &item.Low, &item.Close, &volume, &item.Turnover, &item.Source, &item.DataStatus, &item.CollectedAt); err != nil {
			return nil, err
		}
		item.Timestamp = &stamp
		if volume.Valid {
			item.Volume = int64(volume.Float64)
		}
		items = append(items, item)
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items, rows.Err()
}
