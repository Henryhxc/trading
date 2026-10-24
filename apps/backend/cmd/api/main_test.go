package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ux-securities/backend/internal/config"
	authdomain "ux-securities/backend/internal/domain/auth"
	market "ux-securities/backend/internal/domain/market"
	"ux-securities/backend/internal/httpapi"
	marketservice "ux-securities/backend/internal/service/market"
)

type runtimeProviderFixture struct {
	quote market.Quote
}

type runtimeAuthFixture struct{}

func (runtimeAuthFixture) Register(context.Context, authdomain.Registration) (authdomain.User, string, time.Time, error) {
	return authdomain.User{}, "", time.Time{}, nil
}

func (runtimeAuthFixture) Login(context.Context, authdomain.Credentials) (authdomain.User, string, time.Time, error) {
	return authdomain.User{}, "", time.Time{}, nil
}

func (runtimeAuthFixture) Authenticate(context.Context, string) (authdomain.Actor, error) {
	return authdomain.Actor{UserID: 1, Username: "smoke"}, nil
}

func (runtimeAuthFixture) Logout(context.Context, string) error { return nil }

func (p runtimeProviderFixture) Search(context.Context, string) ([]market.Symbol, error) {
	return []market.Symbol{}, nil
}

func (p runtimeProviderFixture) Quote(context.Context, string) (market.Quote, error) {
	return p.quote, nil
}

func (p runtimeProviderFixture) Candles(context.Context, string, time.Time, time.Time, int) ([]market.Candle, error) {
	return []market.Candle{}, nil
}

func TestBuildMarketDataProviderMockModeIsSimulated(t *testing.T) {
	provider, err := buildMarketDataProvider(config.MarketDataModeMock, nil)
	if err != nil {
		t.Fatalf("buildMarketDataProvider() error = %v, want nil", err)
	}

	quote, err := provider.Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("Quote() error = %v, want nil", err)
	}
	if quote.Source != "mock" || quote.DataStatus != market.DataStatusSimulated {
		t.Fatalf("Quote() source/status = %q/%q, want mock/simulated", quote.Source, quote.DataStatus)
	}
}

func TestBuildMarketDataProviderRepositoryModeIsUnofficial(t *testing.T) {
	price := 380.2
	repository := runtimeProviderFixture{quote: market.Quote{
		Symbol:     "00700",
		LastPrice:  &price,
		Source:     "mysql-cache",
		DataStatus: market.DataStatusLive,
	}}

	provider, err := buildMarketDataProvider(config.MarketDataModeRepository, repository)
	if err != nil {
		t.Fatalf("buildMarketDataProvider() error = %v, want nil", err)
	}
	quote, err := provider.Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("Quote() error = %v, want nil", err)
	}
	if quote.DataStatus != market.DataStatusStale {
		t.Fatalf("repository quote status = %q, want stale safety downgrade", quote.DataStatus)
	}
}

func TestBuildMarketDataProviderRejectsRepositoryModeWithoutRepository(t *testing.T) {
	if _, err := buildMarketDataProvider(config.MarketDataModeRepository, nil); err == nil {
		t.Fatal("buildMarketDataProvider() error = nil, want fail-closed error")
	}
}

func TestBuildMarketDataProviderAkShareModeUsesRepositoryCache(t *testing.T) {
	provider, err := buildMarketDataProvider(config.MarketDataModeAkShare, runtimeProviderFixture{})
	if err != nil {
		t.Fatalf("buildMarketDataProvider() error = %v, want nil", err)
	}
	if provider == nil {
		t.Fatal("buildMarketDataProvider() returned nil provider")
	}
}

func TestBuildMarketServiceUsesConfiguredStaleThreshold(t *testing.T) {
	now := time.Date(2026, 10, 24, 6, 0, 0, 0, time.UTC)
	price := 381.2
	provider := runtimeProviderFixture{quote: market.Quote{Symbol: "00700", LastPrice: &price, AsOf: now.Add(-time.Minute), Source: "akshare", DataStatus: market.DataStatusDelayed}}
	service := buildMarketService(config.Config{MarketDataStaleAfter: 20 * time.Minute}, provider, func() time.Time { return now })
	quote, err := service.Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
	if quote.Source != "akshare" || quote.DataStatus != market.DataStatusDelayed {
		t.Fatalf("quote source/status = %q/%q, want akshare/delayed", quote.Source, quote.DataStatus)
	}
}

func TestMockModeQuoteAPISmoke(t *testing.T) {
	provider, err := buildMarketDataProvider(config.MarketDataModeMock, nil)
	if err != nil {
		t.Fatalf("buildMarketDataProvider() error = %v, want nil", err)
	}
	handler := httpapi.NewWithMarket(runtimeAuthFixture{}, false, marketservice.NewService(provider), nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/markets/quotes/00700", nil)
	request.AddCookie(&http.Cookie{Name: "ux_session", Value: "smoke-session"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET quote status = %d body=%s, want 200", response.Code, response.Body.String())
	}
	var quote market.Quote
	if err := json.NewDecoder(response.Body).Decode(&quote); err != nil {
		t.Fatalf("decode quote response: %v", err)
	}
	if quote.Source != "mock" || quote.DataStatus != market.DataStatusSimulated {
		t.Fatalf("GET quote source/status = %q/%q, want mock/simulated", quote.Source, quote.DataStatus)
	}
}
