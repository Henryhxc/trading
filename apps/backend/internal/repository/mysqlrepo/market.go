package mysqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	market "ux-securities/backend/internal/domain/market"
)

type MarketRepository struct{ db *sql.DB }

func NewMarketRepository(db *sql.DB) *MarketRepository { return &MarketRepository{db: db} }

func (r *MarketRepository) Search(ctx context.Context, query string) ([]market.Symbol, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+cachedQuoteColumns+" FROM stocks WHERE symbol LIKE CONCAT('%', ?, '%') OR name LIKE CONCAT('%', ?, '%') ORDER BY symbol", query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []market.Symbol{}
	for rows.Next() {
		quote, err := scanCachedQuote(rows)
		if err != nil {
			return nil, err
		}
		item := market.Symbol{Symbol: quote.Symbol, DisplayName: quote.DisplayName}
		if quote.LastPrice != nil {
			item.Quote = &quote
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MarketRepository) Quote(ctx context.Context, symbol string) (market.Quote, error) {
	q, err := scanCachedQuote(r.db.QueryRowContext(ctx, "SELECT "+cachedQuoteColumns+" FROM stocks WHERE symbol = ?", symbol))
	if errors.Is(err, sql.ErrNoRows) {
		return q, market.ErrNotFound
	}
	return q, err
}

const cachedQuoteColumns = "symbol, name, last_price, bid, ask, `change`, change_percent, volume, turnover, quote_as_of, quote_source, data_status"

func scanCachedQuote(row interface{ Scan(...any) error }) (market.Quote, error) {
	var q market.Quote
	var last, bid, ask, change, percent, volume, turnover sql.NullFloat64
	var asOf sql.NullTime
	var source sql.NullString
	err := row.Scan(&q.Symbol, &q.DisplayName, &last, &bid, &ask, &change, &percent, &volume, &turnover, &asOf, &source, &q.DataStatus)
	if err != nil {
		return q, err
	}
	q.LastPrice, q.Bid, q.Ask, q.Change, q.ChangePercent = nullableFloat(last), nullableFloat(bid), nullableFloat(ask), nullableFloat(change), nullableFloat(percent)
	if volume.Valid {
		q.Volume = int64(volume.Float64)
	}
	q.AsOf = asOf.Time
	q.Turnover, q.Source = turnover.Float64, source.String
	if !last.Valid {
		q.DataStatus = market.DataStatusUnavailable
	}
	if !source.Valid {
		q.Source = "none"
	}
	return q, nil
}

func (r *MarketRepository) Candles(ctx context.Context, symbol string, from, to time.Time, limit int) ([]market.Candle, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx, "SELECT 1 FROM stocks WHERE symbol = ?", symbol).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return nil, market.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	query := "SELECT q.trade_date, q.open, q.high, q.low, q.close, q.volume, q.turnover, q.source, q.data_status, q.collected_at FROM stock_daily_quotes q JOIN stocks s ON s.id = q.stock_id WHERE s.symbol = ?"
	args := []any{symbol}
	if !from.IsZero() {
		query += " AND q.trade_date >= ?"
		args = append(args, from)
	}
	if !to.IsZero() {
		query += " AND q.trade_date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY q.trade_date DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []market.Candle{}
	for rows.Next() {
		var item market.Candle
		var volume sql.NullFloat64
		if err := rows.Scan(&item.TradeDate, &item.Open, &item.High, &item.Low, &item.Close, &volume, &item.Turnover, &item.Source, &item.DataStatus, &item.CollectedAt); err != nil {
			return nil, err
		}
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

func nullableFloat(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}
