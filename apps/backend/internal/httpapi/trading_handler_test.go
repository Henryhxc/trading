package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	auth "ux-securities/backend/internal/domain/auth"
	domain "ux-securities/backend/internal/domain/trading"
	"ux-securities/backend/internal/service/trading"
)

type tradingHandlerFixture struct{}

func (tradingHandlerFixture) Create(context.Context, auth.Actor, trading.CreateOrderInput) (trading.Order, error) {
	return trading.Order{ID: 1, Symbol: "00700", Status: domain.StatusFilled}, nil
}
func (tradingHandlerFixture) Cancel(context.Context, auth.Actor, int64) (trading.Order, error) {
	return trading.Order{}, nil
}
func (tradingHandlerFixture) Order(context.Context, auth.Actor, int64) (trading.Order, error) {
	return trading.Order{}, nil
}
func (tradingHandlerFixture) Account(context.Context, auth.Actor) (trading.Account, error) {
	return trading.Account{Currency: "HKD", CashBalance: 1000000, AvailableCash: 1000000, InitialCash: 1000000, Status: "active"}, nil
}
func (tradingHandlerFixture) Positions(context.Context, auth.Actor) ([]trading.Position, error) {
	return nil, nil
}
func (tradingHandlerFixture) Orders(context.Context, auth.Actor, trading.OrderFilter) ([]trading.Order, error) {
	return nil, nil
}

type tradingHandlerAuth struct{}

func (tradingHandlerAuth) Register(context.Context, auth.Registration) (auth.User, string, time.Time, error) {
	return auth.User{}, "", time.Time{}, nil
}
func (tradingHandlerAuth) Login(context.Context, auth.Credentials) (auth.User, string, time.Time, error) {
	return auth.User{}, "", time.Time{}, nil
}
func (tradingHandlerAuth) Authenticate(context.Context, string) (auth.Actor, error) {
	return auth.Actor{UserID: 1}, nil
}
func (tradingHandlerAuth) Logout(context.Context, string) error { return nil }

type idempotencyHeaderFixture struct{ clientOrderID string }

func (f *idempotencyHeaderFixture) Create(_ context.Context, _ auth.Actor, input trading.CreateOrderInput) (trading.Order, error) {
	f.clientOrderID = input.ClientOrderID
	return trading.Order{ID: 1, Symbol: "00700", Status: domain.StatusFilled}, nil
}
func (*idempotencyHeaderFixture) Cancel(context.Context, auth.Actor, int64) (trading.Order, error) {
	return trading.Order{}, nil
}
func (*idempotencyHeaderFixture) Order(context.Context, auth.Actor, int64) (trading.Order, error) {
	return trading.Order{}, nil
}
func (*idempotencyHeaderFixture) Account(context.Context, auth.Actor) (trading.Account, error) {
	return trading.Account{}, nil
}
func (*idempotencyHeaderFixture) Positions(context.Context, auth.Actor) ([]trading.Position, error) {
	return nil, nil
}
func (*idempotencyHeaderFixture) Orders(context.Context, auth.Actor, trading.OrderFilter) ([]trading.Order, error) {
	return nil, nil
}

func TestCreateOrderAcceptsLimitOrderTypeFromFrontend(t *testing.T) {
	handler := NewWithTrading(tradingHandlerAuth{}, false, nil, nil, tradingHandlerFixture{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(`{"client_order_id":"client-12345","symbol":"00700","side":"buy","order_type":"limit","quantity":100,"limit_price":380.2}`))
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST /orders status = %d body=%s, want 201", response.Code, response.Body.String())
	}
}

func TestCreateOrderUsesIdempotencyKeyHeader(t *testing.T) {
	fixture := &idempotencyHeaderFixture{}
	handler := NewWithTrading(tradingHandlerAuth{}, false, nil, nil, fixture)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(`{"symbol":"00700","side":"buy","order_type":"limit","quantity":100,"limit_price":380.2}`))
	request.Header.Set("Idempotency-Key", "client-order-0001")
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST /orders status = %d body=%s, want 201", response.Code, response.Body.String())
	}
	if fixture.clientOrderID != "client-order-0001" {
		t.Fatalf("service idempotency key = %q, want header value", fixture.clientOrderID)
	}
}

func TestAccountResponseIncludesInitialCashAndStatus(t *testing.T) {
	handler := NewWithTrading(tradingHandlerAuth{}, false, nil, nil, tradingHandlerFixture{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/account", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /account status = %d body=%s, want 200", response.Code, response.Body.String())
	}
	var account map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &account); err != nil {
		t.Fatalf("decode account response: %v", err)
	}
	for _, field := range []string{"initial_cash", "status"} {
		if _, exists := account[field]; !exists {
			t.Errorf("GET /account response missing %q: %s", field, response.Body.String())
		}
	}
}
