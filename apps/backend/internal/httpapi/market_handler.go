package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	domain "ux-securities/backend/internal/domain/auth"
	market "ux-securities/backend/internal/domain/market"
)

func (s *server) searchSymbols(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if len(query) > 80 {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Check the submitted details")
		return
	}
	items, err := s.market.Search(r.Context(), query)
	if err != nil {
		s.fail(w, r, http.StatusServiceUnavailable, "market_unavailable", "Market data is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *server) getQuote(w http.ResponseWriter, r *http.Request) {
	quote, err := s.market.Quote(r.Context(), r.PathValue("symbol"))
	if err != nil {
		if errors.Is(err, market.ErrUnavailable) {
			s.fail(w, r, http.StatusServiceUnavailable, "market_unavailable", "Market data is unavailable")
			return
		}
		if errors.Is(err, market.ErrNotFound) {
			s.fail(w, r, http.StatusNotFound, "not_found", "Security was not found")
			return
		}
		s.fail(w, r, http.StatusNotFound, "not_found", "Security was not found")
		return
	}
	if quote.DataStatus == market.DataStatusUnavailable {
		s.fail(w, r, http.StatusServiceUnavailable, "market_unavailable", "Market data is unavailable")
		return
	}
	_ = quote.DataStatus // wire contract includes the data_status field
	writeJSON(w, http.StatusOK, quote)
}

func (s *server) getDailyCandles(w http.ResponseWriter, r *http.Request) {
	if _, provided := r.URL.Query()["interval"]; provided {
		s.getCandleSeries(w, r)
		return
	}
	from, fromProvided, validFrom := parseOptionalCandleDate(r.URL.Query(), "from")
	to, toProvided, validTo := parseOptionalCandleDate(r.URL.Query(), "to")
	limit, validLimit := parseCandleLimit(r.URL.Query())
	if !validFrom || !validTo || !validLimit || (fromProvided && toProvided && from.After(to)) {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Check the submitted details")
		return
	}
	items, err := s.market.Candles(r.Context(), r.PathValue("symbol"), from, to, limit)
	if err != nil {
		if errors.Is(err, market.ErrUnavailable) {
			s.fail(w, r, http.StatusServiceUnavailable, "market_unavailable", "Market data is unavailable")
			return
		}
		if errors.Is(err, market.ErrNotFound) {
			s.fail(w, r, http.StatusNotFound, "not_found", "Security was not found")
			return
		}
		s.fail(w, r, http.StatusNotFound, "not_found", "Security was not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"symbol": normalizeSymbol(r.PathValue("symbol")), "interval": "1d", "items": items})
}

func parseOptionalCandleDate(values url.Values, name string) (time.Time, bool, bool) {
	raw, provided := values[name]
	if !provided {
		return time.Time{}, false, true
	}
	if len(raw) != 1 {
		return time.Time{}, true, false
	}
	value, err := time.Parse("2006-01-02", raw[0])
	if err != nil {
		return time.Time{}, true, false
	}
	return value, true, true
}

func parseCandleLimit(values url.Values) (int, bool) {
	raw, provided := values["limit"]
	if !provided {
		return 90, true
	}
	if len(raw) != 1 {
		return 0, false
	}
	limit, err := strconv.Atoi(raw[0])
	if err != nil || limit < 1 || limit > 365 {
		return 0, false
	}
	return limit, true
}

func (s *server) listWatchlist(w http.ResponseWriter, r *http.Request) {
	items, err := s.watchlist.List(r.Context(), actorFrom(r))
	if err != nil {
		s.fail(w, r, http.StatusInternalServerError, "internal_error", "Request could not be completed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (s *server) addWatchlist(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Symbol string `json:"symbol"`
	}
	if !decode(w, r, &body) || !validSymbol(body.Symbol) {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Check the submitted details")
		return
	}
	item, created, err := s.watchlist.Add(r.Context(), actorFrom(r), body.Symbol)
	if err != nil {
		s.fail(w, r, http.StatusNotFound, "not_found", "Security was not found")
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, item)
}
func (s *server) deleteWatchlist(w http.ResponseWriter, r *http.Request) {
	if err := s.watchlist.Delete(r.Context(), actorFrom(r), r.PathValue("symbol")); err != nil {
		s.fail(w, r, http.StatusInternalServerError, "internal_error", "Request could not be completed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *server) reorderWatchlist(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Symbols []string `json:"symbols"`
	}
	if !decode(w, r, &body) {
		s.fail(w, r, http.StatusBadRequest, "invalid_request", "Request body is invalid")
		return
	}
	if err := s.watchlist.Reorder(r.Context(), actorFrom(r), body.Symbols); err != nil {
		s.fail(w, r, http.StatusInternalServerError, "internal_error", "Request could not be completed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func actorFrom(r *http.Request) domain.Actor { return r.Context().Value(actorKey{}).(domain.Actor) }
func normalizeSymbol(value string) string {
	value = strings.TrimSpace(value)
	for len(value) < 5 {
		value = "0" + value
	}
	return value
}
func validSymbol(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 5 || len(value) > 6 {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
