package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdomain "ux-securities/backend/internal/domain/auth"
	marketdomain "ux-securities/backend/internal/domain/market"
	watchlistdomain "ux-securities/backend/internal/domain/watchlist"
)

type marketHandlerMarketStub struct {
	searchCalls int
	candleCalls int
	candleFrom  time.Time
	candleTo    time.Time
	candleLimit int
}

type marketUnavailableStub struct{}

func (*marketUnavailableStub) Search(context.Context, string) ([]marketdomain.Symbol, error) {
	return nil, marketdomain.ErrUnavailable
}
func (*marketUnavailableStub) Quote(context.Context, string) (marketdomain.Quote, error) {
	return marketdomain.Quote{}, marketdomain.ErrUnavailable
}
func (*marketUnavailableStub) Candles(context.Context, string, time.Time, time.Time, int) ([]marketdomain.Candle, error) {
	return nil, marketdomain.ErrUnavailable
}

func (s *marketHandlerMarketStub) Search(context.Context, string) ([]marketdomain.Symbol, error) {
	s.searchCalls++
	return []marketdomain.Symbol{{Symbol: "00700", DisplayName: "Tencent Holdings"}}, nil
}
func (*marketHandlerMarketStub) Quote(context.Context, string) (marketdomain.Quote, error) {
	return marketdomain.Quote{}, nil
}
func (s *marketHandlerMarketStub) Candles(_ context.Context, _ string, from, to time.Time, limit int) ([]marketdomain.Candle, error) {
	s.candleCalls++
	s.candleFrom = from
	s.candleTo = to
	s.candleLimit = limit
	return nil, nil
}

type marketHandlerWatchlistStub struct {
	created bool
}

func (*marketHandlerWatchlistStub) List(context.Context, authdomain.Actor) ([]watchlistdomain.Item, error) {
	return nil, nil
}
func (s *marketHandlerWatchlistStub) Add(context.Context, authdomain.Actor, string) (watchlistdomain.Item, bool, error) {
	return watchlistdomain.Item{Symbol: "00700", Name: "Tencent Holdings"}, s.created, nil
}
func (*marketHandlerWatchlistStub) Delete(context.Context, authdomain.Actor, string) error {
	return nil
}
func (*marketHandlerWatchlistStub) Reorder(context.Context, authdomain.Actor, []string) error {
	return nil
}

func authenticatedRequest(method, target string, body *string) *http.Request {
	var requestBody io.Reader
	if body != nil {
		requestBody = strings.NewReader(*body)
	}
	request := httptest.NewRequest(method, target, requestBody)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid-session"})
	return request
}

func TestEmptySearchListsDirectory(t *testing.T) {
	market := &marketHandlerMarketStub{}
	handler := NewWithMarket(&testAuthService{}, false, market, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/markets/symbols?query=", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("directory status = %d, want 200", response.Code)
	}
	if market.searchCalls != 1 {
		t.Fatalf("directory called provider %d times, want 1", market.searchCalls)
	}
}

func TestAddWatchlistUsesCreatedStatusOnlyForNewRelation(t *testing.T) {
	for _, test := range []struct {
		name    string
		created bool
		want    int
	}{
		{name: "new", created: true, want: http.StatusCreated},
		{name: "duplicate", created: false, want: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			watchlist := &marketHandlerWatchlistStub{created: test.created}
			handler := NewWithMarket(&testAuthService{}, false, nil, watchlist)
			response := httptest.NewRecorder()
			body := `{"symbol":"00700"}`
			request := httptest.NewRequest(http.MethodPost, "/api/v1/watchlist", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid-session"})
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("created=%v status = %d, want %d", test.created, response.Code, test.want)
			}
		})
	}
}

func TestGetDailyCandlesRejectsInvalidQueryBeforeProviderCall(t *testing.T) {
	for _, query := range []string{
		"from=2026-10-02x",
		"to=2026-10-01x",
		"from=2026-10-02&to=2026-10-01",
		"limit=0",
		"limit=366",
		"limit=not-a-number",
	} {
		t.Run(query, func(t *testing.T) {
			market := &marketHandlerMarketStub{}
			handler := NewWithMarket(&testAuthService{}, false, market, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/markets/candles/00700?"+query, nil))
			if response.Code != http.StatusBadRequest {
				t.Fatalf("query %q status = %d, want 400", query, response.Code)
			}
			if market.candleCalls != 0 {
				t.Fatalf("query %q called provider %d times, want 0", query, market.candleCalls)
			}
		})
	}
}

func TestGetDailyCandlesUsesValidatedDateRangeAndDefaultLimit(t *testing.T) {
	market := &marketHandlerMarketStub{}
	handler := NewWithMarket(&testAuthService{}, false, market, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/markets/candles/00700?from=2026-10-01&to=2026-10-24", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("valid query status = %d, want 200", response.Code)
	}
	if market.candleCalls != 1 || market.candleLimit != 90 {
		t.Fatalf("provider call = %d, limit = %d, want one call with default limit 90", market.candleCalls, market.candleLimit)
	}
	if market.candleFrom.Format("2006-01-02") != "2026-10-01" || market.candleTo.Format("2006-01-02") != "2026-10-24" {
		t.Fatalf("provider date range = %s..%s, want 2026-10-01..2026-10-24", market.candleFrom.Format("2006-01-02"), market.candleTo.Format("2006-01-02"))
	}
}

func TestGetQuoteMapsUnavailableProviderErrorToServiceUnavailable(t *testing.T) {
	handler := NewWithMarket(&testAuthService{}, false, &marketUnavailableStub{}, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/markets/quotes/00700", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unavailable quote status = %d, want 503", response.Code)
	}
}

func TestGetDailyCandlesMapsUnavailableProviderErrorToServiceUnavailable(t *testing.T) {
	handler := NewWithMarket(&testAuthService{}, false, &marketUnavailableStub{}, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/v1/markets/candles/00700", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unavailable candles status = %d, want 503", response.Code)
	}
}

type marketNotFoundStub struct{}

func (*marketNotFoundStub) Search(context.Context, string) ([]marketdomain.Symbol, error) {
	return nil, marketdomain.ErrNotFound
}
func (*marketNotFoundStub) Quote(context.Context, string) (marketdomain.Quote, error) {
	return marketdomain.Quote{}, marketdomain.ErrNotFound
}
func (*marketNotFoundStub) Candles(context.Context, string, time.Time, time.Time, int) ([]marketdomain.Candle, error) {
	return nil, marketdomain.ErrNotFound
}

func TestMarketHandlersMapUnknownSymbolToNotFound(t *testing.T) {
	handler := NewWithMarket(&testAuthService{}, false, &marketNotFoundStub{}, nil)
	for _, target := range []string{
		"/api/v1/markets/quotes/12345",
		"/api/v1/markets/candles/12345",
	} {
		t.Run(target, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusNotFound {
				t.Fatalf("unknown symbol status = %d, want 404", response.Code)
			}
		})
	}
}
