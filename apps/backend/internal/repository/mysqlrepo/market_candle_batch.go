package mysqlrepo

import (
	"context"
	"database/sql"
	"strings"
	"time"

	domain "ux-securities/backend/internal/domain/market"
)

type candleBatchKey struct {
	symbol, interval, source string
	status                   domain.DataStatus
	collected                time.Time
}

func upsertCandles(ctx context.Context, tx *sql.Tx, candles []domain.CandleUpdate) (map[string]domain.MarketEvent, error) {
	groups := make(map[candleBatchKey][]domain.CandleUpdate)
	keys := []candleBatchKey{}
	for _, candle := range candles {
		if candle.Interval == "" {
			candle.Interval = "1d"
		}
		key := candleBatchKey{candle.Symbol, candle.Interval, candle.Source, candle.DataStatus, candle.CollectedAt}
		if _, exists := groups[key]; !exists {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], candle)
	}
	ids := make(map[string]uint64)
	events := make(map[string]domain.MarketEvent)
	for _, key := range keys {
		stockID, exists := ids[key.symbol]
		if !exists {
			// History creates metadata once per symbol, without inventing a spot quote.
			if _, err := tx.ExecContext(ctx, `INSERT INTO stocks (symbol, name, market, status, volume, turnover, quote_source, data_status, updated_at)
			 VALUES (?, ?, 'HKEX', 'active', 0, 0, 'none', 'unavailable', ?) ON DUPLICATE KEY UPDATE symbol=VALUES(symbol)`, key.symbol, key.symbol, key.collected); err != nil {
				return nil, err
			}
			if err := tx.QueryRowContext(ctx, "SELECT id FROM stocks WHERE symbol = ?", key.symbol).Scan(&stockID); err != nil {
				return nil, err
			}
			ids[key.symbol] = stockID
		}
		group := groups[key]
		for start := 0; start < len(group); start += 400 {
			end := min(start+400, len(group))
			changed, err := upsertCandleRows(ctx, tx, stockID, group[start:end])
			if err != nil {
				return nil, err
			}
			eventKey := key.symbol + ":" + key.interval
			if changed && (events[eventKey].AsOf.IsZero() || !events[eventKey].AsOf.After(key.collected)) {
				events[eventKey] = domain.MarketEvent{Type: "candles.updated", Symbol: key.symbol, Interval: key.interval, Source: key.source, DataStatus: key.status, AsOf: key.collected}
			}
		}
	}
	return events, nil
}

func upsertCandleRows(ctx context.Context, tx *sql.Tx, stockID uint64, candles []domain.CandleUpdate) (bool, error) {
	table, columns, placeholders := "stock_daily_quotes", "stock_id, trade_date", "?, ?"
	minutes := domain.IsMinuteInterval(candles[0].Interval)
	if minutes {
		table, columns, placeholders = "stock_intraday_quotes", "stock_id, trade_date, interval_code, bar_time", "?, ?, ?, ?"
	}
	columns += ", `open`, high, low, close, volume, turnover, source, data_status, collected_at, adjustment"
	placeholders = "(" + placeholders + ", ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	values := make([]string, 0, len(candles))
	args := make([]any, 0, len(candles)*14)
	for _, candle := range candles {
		values = append(values, placeholders)
		args = append(args, stockID, candle.TradeDate)
		if minutes {
			args = append(args, candle.Interval, candle.Timestamp.UTC())
		}
		var turnover any = candle.Turnover
		if candle.Turnover == "" {
			turnover = nil
		}
		args = append(args, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, turnover, candle.Source, candle.DataStatus, candle.CollectedAt, candle.Adjustment)
	}
	updates := ""
	for _, field := range []string{"`open`", "high", "low", "close", "volume", "turnover", "source", "data_status", "adjustment"} {
		updates += field + " = IF(VALUES(collected_at) > collected_at, VALUES(" + field + "), " + field + "), "
	}
	updates += "collected_at = GREATEST(collected_at, VALUES(collected_at))"
	// Only fixed identifiers and placeholders are assembled. All market values are bound.
	result, err := tx.ExecContext(ctx, "INSERT INTO "+table+" ("+columns+") VALUES "+strings.Join(values, ",")+" ON DUPLICATE KEY UPDATE "+updates, args...)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows > 0, err
}
