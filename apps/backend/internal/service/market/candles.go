package market

import (
	"context"
	"sort"
	"time"
	domain "ux-securities/backend/internal/domain/market"
)

var hongKong = time.FixedZone("HKT", 8*60*60)

func (s *Service) CandleSeries(ctx context.Context, symbol, interval string, from, to time.Time, limit int) (domain.CandleSeries, error) {
	series := domain.CandleSeries{Symbol: symbol, Interval: interval, Items: []domain.Candle{}, Source: "none", DataStatus: domain.DataStatusUnavailable}
	var items []domain.Candle
	var err error
	if domain.IsMinuteInterval(interval) {
		provider, ok := s.provider.(domain.NativeCandleProvider)
		if !ok {
			return series, nil
		}
		items, err = provider.NativeCandles(ctx, symbol, interval, from, to, limit)
	} else {
		// At most 366 calendar days are accepted. Aggregate before applying the
		// output limit, so a weekly/monthly limit does not truncate its source days.
		items, err = s.provider.Candles(ctx, symbol, from, to, 366)
	}
	if err != nil {
		return series, err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Timestamp != nil && items[j].Timestamp != nil {
			return items[i].Timestamp.Before(*items[j].Timestamp)
		}
		return items[i].TradeDate.Before(items[j].TradeDate)
	})
	if domain.IsMinuteInterval(interval) {
		for i := range items {
			if isUnofficialStatus(items[i].DataStatus) && s.now().Sub(items[i].CollectedAt) > s.staleAfter {
				items[i].DataStatus = domain.DataStatusStale
			}
		}
	}
	if !domain.IsMinuteInterval(interval) && interval != "1d" {
		items = aggregateCandles(items, interval)
	}
	if len(items) > limit {
		items = items[len(items)-limit:]
	}
	if len(items) == 0 {
		return series, nil
	}
	series.Items = items
	series.Source, series.DataStatus = items[0].Source, items[0].DataStatus
	var asOf time.Time
	for _, bar := range items {
		if bar.Source != series.Source {
			series.Source = "mixed"
		}
		if statusSeverity(bar.DataStatus) > statusSeverity(series.DataStatus) {
			series.DataStatus = bar.DataStatus
		}
		if bar.CollectedAt.After(asOf) {
			asOf = bar.CollectedAt
		}
	}
	if !asOf.IsZero() {
		series.AsOf = &asOf
	}
	return series, nil
}

func aggregateCandles(items []domain.Candle, interval string) []domain.Candle {
	result := []domain.Candle{}
	for _, bar := range items {
		// Daily dates are exchange-calendar labels, not UTC instants to shift.
		y, m, d := bar.TradeDate.Date()
		bucket := time.Date(y, m, d, 0, 0, 0, 0, hongKong)
		switch interval {
		case "1w":
			bucket = bucket.AddDate(0, 0, -(int(bucket.Weekday())+6)%7)
		case "1M":
			bucket = time.Date(y, m, 1, 0, 0, 0, 0, hongKong)
		case "1y":
			bucket = time.Date(y, 1, 1, 0, 0, 0, 0, hongKong)
		}
		if len(result) == 0 || !result[len(result)-1].TradeDate.Equal(bucket) {
			bar.TradeDate = bucket
			bar.Timestamp = nil
			result = append(result, bar)
			continue
		}
		last := &result[len(result)-1]
		if bar.High > last.High {
			last.High = bar.High
		}
		if bar.Low < last.Low {
			last.Low = bar.Low
		}
		last.Close = bar.Close
		last.Volume += bar.Volume
		if last.Turnover != nil && bar.Turnover != nil {
			amount := *last.Turnover + *bar.Turnover
			last.Turnover = &amount
		} else {
			last.Turnover = nil
		}
		if bar.Source != last.Source {
			last.Source = "mixed"
		}
		if statusSeverity(bar.DataStatus) > statusSeverity(last.DataStatus) {
			last.DataStatus = bar.DataStatus
		}
		if bar.CollectedAt.After(last.CollectedAt) {
			last.CollectedAt = bar.CollectedAt
		}
	}
	return result
}

func statusSeverity(status domain.DataStatus) int {
	switch status {
	case domain.DataStatusUnavailable:
		return 5
	case domain.DataStatusSimulated:
		return 4
	case domain.DataStatusStale:
		return 3
	case domain.DataStatusDelayed:
		return 2
	case domain.DataStatusUnofficial:
		return 1
	}
	return 0
}
