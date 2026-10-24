package market

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	domain "ux-securities/backend/internal/domain/market"
)

var (
	ErrInvalidIngestion = errors.New("invalid market ingestion")
	symbolPattern       = regexp.MustCompile(`^[0-9]{5}$`)
	decimalPattern      = regexp.MustCompile(`^-?[0-9]+(?:\.[0-9]{1,4})?$`)
)

type IngestionRepository interface {
	UpsertMarketBatch(context.Context, domain.IngestionBatch) (domain.IngestionResult, error)
}

type MarketPublisher interface {
	Publish(domain.MarketEvent)
}

type IngestionService struct {
	repository IngestionRepository
	publisher  MarketPublisher
}

func NewIngestionService(repository IngestionRepository, publisher MarketPublisher) *IngestionService {
	return &IngestionService{repository: repository, publisher: publisher}
}

func (s *IngestionService) Ingest(ctx context.Context, batch domain.IngestionBatch) (domain.IngestionResult, error) {
	if err := validateIngestionBatch(batch); err != nil {
		return domain.IngestionResult{}, err
	}
	result, err := s.repository.UpsertMarketBatch(ctx, batch)
	if err != nil {
		return domain.IngestionResult{}, err
	}
	result.QuoteCount = len(batch.Quotes)
	result.CandleCount = len(batch.Candles)
	for _, event := range result.Events {
		s.publisher.Publish(event)
	}
	return result, nil
}

func validateIngestionBatch(batch domain.IngestionBatch) error {
	if batch.Version != 1 || len(batch.Quotes)+len(batch.Candles) == 0 {
		return ErrInvalidIngestion
	}
	for _, quote := range batch.Quotes {
		if !symbolPattern.MatchString(quote.Symbol) || strings.TrimSpace(quote.DisplayName) == "" ||
			quote.Source != "akshare" || quote.DataStatus != domain.DataStatusDelayed || quote.CollectedAt.IsZero() ||
			!validNonNegativeDecimal(quote.LastPrice) || !validDecimal(quote.Change) ||
			!validDecimal(quote.ChangePercent) || !validNonNegativeDecimal(quote.Volume) ||
			!validNonNegativeDecimal(quote.Turnover) || !validOptionalNonNegativeDecimal(quote.Bid) ||
			!validOptionalNonNegativeDecimal(quote.Ask) {
			return ErrInvalidIngestion
		}
	}
	seenCandles := make(map[string]struct{}, len(batch.Candles))
	for _, candle := range batch.Candles {
		if candle.Interval == "" {
			candle.Interval = "1d"
		}
		key := candle.Symbol + ":" + candle.Interval + ":" + candle.TradeDate
		if candle.Timestamp != nil {
			key = candle.Symbol + ":" + candle.Interval + ":" + candle.Timestamp.UTC().Format(time.RFC3339Nano)
		}
		if _, duplicate := seenCandles[key]; duplicate {
			return ErrInvalidIngestion
		}
		seenCandles[key] = struct{}{}
		if !validCandle(candle) {
			return ErrInvalidIngestion
		}
	}
	return nil
}

func validCandle(candle domain.CandleUpdate) bool {
	if !symbolPattern.MatchString(candle.Symbol) || (candle.Source != "akshare" && candle.Source != "tencent" && candle.Source != "eastmoney" && candle.Source != "yahoo") ||
		candle.DataStatus != domain.DataStatusUnofficial || candle.CollectedAt.IsZero() ||
		candle.Adjustment != "none" {
		return false
	}
	if _, err := time.Parse("2006-01-02", candle.TradeDate); err != nil {
		return false
	}
	if domain.IsMinuteInterval(candle.Interval) {
		if candle.Timestamp == nil || candle.Timestamp.IsZero() || candle.Timestamp.In(hongKong).Format("2006-01-02") != candle.TradeDate {
			return false
		}
	} else if (candle.Interval != "" && candle.Interval != "1d") || candle.Timestamp != nil {
		return false
	}
	open, openOK := decimalValue(candle.Open, false)
	high, highOK := decimalValue(candle.High, false)
	low, lowOK := decimalValue(candle.Low, false)
	closeValue, closeOK := decimalValue(candle.Close, false)
	missingAmount := candle.Turnover == "" && ((candle.Source == "tencent" && (candle.Interval == "" || candle.Interval == "1d")) || candle.Source == "yahoo")
	if !openOK || !highOK || !lowOK || !closeOK ||
		!validNonNegativeDecimal(candle.Volume) || (!missingAmount && !validNonNegativeDecimal(candle.Turnover)) {
		return false
	}
	return low.Cmp(open) <= 0 && low.Cmp(closeValue) <= 0 && high.Cmp(open) >= 0 && high.Cmp(closeValue) >= 0
}

func validOptionalNonNegativeDecimal(value *string) bool {
	return value == nil || validNonNegativeDecimal(*value)
}

func validNonNegativeDecimal(value string) bool {
	_, ok := decimalValue(value, false)
	return ok
}

func validDecimal(value string) bool {
	_, ok := decimalValue(value, true)
	return ok
}

func decimalValue(value string, allowNegative bool) (*big.Rat, bool) {
	if !decimalPattern.MatchString(value) {
		return nil, false
	}
	number, ok := new(big.Rat).SetString(value)
	if !ok || (!allowNegative && number.Sign() < 0) {
		return nil, false
	}
	return number, true
}
