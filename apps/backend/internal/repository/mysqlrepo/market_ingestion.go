package mysqlrepo

import (
	"context"
	"database/sql"
	"strconv"

	domain "ux-securities/backend/internal/domain/market"
)

func (r *MarketRepository) UpsertMarketBatch(ctx context.Context, batch domain.IngestionBatch) (domain.IngestionResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.IngestionResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	events := make([]domain.MarketEvent, 0, len(batch.Quotes)+len(batch.Candles))
	for _, quote := range batch.Quotes {
		changed, err := upsertQuote(ctx, tx, quote)
		if err != nil {
			return domain.IngestionResult{}, err
		}
		if changed {
			events = append(events, quoteEvent(quote))
		}
	}
	candleEvents, err := upsertCandles(ctx, tx, batch.Candles)
	if err != nil {
		return domain.IngestionResult{}, err
	}
	for _, event := range candleEvents {
		events = append(events, event)
	}
	if err := tx.Commit(); err != nil {
		return domain.IngestionResult{}, err
	}
	return domain.IngestionResult{Events: events}, nil
}

func upsertQuote(ctx context.Context, tx *sql.Tx, quote domain.QuoteUpdate) (bool, error) {
	result, err := tx.ExecContext(ctx, `INSERT INTO stocks
		(symbol, name, market, status, last_price, bid, ask, `+"`change`"+`, change_percent, volume, turnover, quote_as_of, quote_source, data_status, updated_at)
		VALUES (?, ?, 'HKEX', 'active', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		name = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(name), name),
		last_price = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(last_price), last_price),
		bid = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(bid), bid),
		ask = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(ask), ask),
		`+"`change`"+` = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(`+"`change`"+`), `+"`change`"+`),
		change_percent = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(change_percent), change_percent),
		volume = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(volume), volume),
		turnover = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(turnover), turnover),
		quote_source = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(quote_source), quote_source),
		data_status = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(data_status), data_status),
		updated_at = IF(quote_as_of IS NULL OR VALUES(quote_as_of) > quote_as_of, VALUES(updated_at), updated_at),
		quote_as_of = GREATEST(COALESCE(quote_as_of, VALUES(quote_as_of)), VALUES(quote_as_of))`,
		quote.Symbol, quote.DisplayName, quote.LastPrice, quote.Bid, quote.Ask,
		quote.Change, quote.ChangePercent, quote.Volume, quote.Turnover,
		quote.CollectedAt, quote.Source, quote.DataStatus, quote.CollectedAt,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows > 0, err
}

func quoteEvent(update domain.QuoteUpdate) domain.MarketEvent {
	quote := domain.Quote{
		Symbol: update.Symbol, DisplayName: update.DisplayName,
		LastPrice: decimalFloat(update.LastPrice), Bid: optionalDecimalFloat(update.Bid), Ask: optionalDecimalFloat(update.Ask),
		Change: decimalFloat(update.Change), ChangePercent: decimalFloat(update.ChangePercent),
		Volume: int64Value(update.Volume), Turnover: floatValue(update.Turnover),
		AsOf: update.CollectedAt, Source: update.Source, DataStatus: update.DataStatus,
	}
	return domain.MarketEvent{Type: "quote.updated", Symbol: update.Symbol, Source: update.Source, DataStatus: update.DataStatus, AsOf: update.CollectedAt, Quote: &quote}
}

func decimalFloat(value string) *float64 {
	parsed := floatValue(value)
	return &parsed
}

func optionalDecimalFloat(value *string) *float64 {
	if value == nil {
		return nil
	}
	return decimalFloat(*value)
}

func floatValue(value string) float64 {
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func int64Value(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}
