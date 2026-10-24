package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
	domain "ux-securities/backend/internal/domain/market"
)

type candleSeriesAPI interface {
	CandleSeries(context.Context, string, string, time.Time, time.Time, int) (domain.CandleSeries, error)
}

func (s *server) getCandleSeries(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	intervals := values["interval"]
	from, hasFrom, validFrom := parseOptionalCandleDate(values, "from")
	to, hasTo, validTo := parseOptionalCandleDate(values, "to")
	limit := 5000
	validLimit := true
	if raw, exists := values["limit"]; exists {
		if len(raw) != 1 {
			validLimit = false
		} else {
			var err error
			limit, err = strconv.Atoi(raw[0])
			validLimit = err == nil && limit > 0 && limit <= 5000
		}
	}
	if !hasTo {
		now := time.Now().In(time.FixedZone("HKT", 8*60*60))
		to = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
	if !hasFrom {
		monthStart := time.Date(to.Year(), to.Month()-6, 1, 0, 0, 0, 0, time.UTC)
		day := to.Day()
		if last := monthStart.AddDate(0, 1, -1).Day(); day > last {
			day = last
		}
		from = monthStart.AddDate(0, 0, day-1)
	}
	if len(intervals) != 1 || !domain.IsCandleInterval(intervals[0]) || !validFrom || !validTo || !validLimit || from.After(to) || to.Sub(from) > 366*24*time.Hour {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Choose a valid interval and a date window of at most 366 days")
		return
	}
	provider, ok := s.market.(candleSeriesAPI)
	if !ok {
		s.fail(w, r, http.StatusServiceUnavailable, "market_unavailable", "Candle service is unavailable")
		return
	}
	series, err := provider.CandleSeries(r.Context(), normalizeSymbol(r.PathValue("symbol")), intervals[0], from, to, limit)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.fail(w, r, http.StatusNotFound, "not_found", "Security was not found")
			return
		}
		s.fail(w, r, http.StatusServiceUnavailable, "market_unavailable", "Candle data is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, series)
}
