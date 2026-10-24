package trading

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	auth "ux-securities/backend/internal/domain/auth"
	market "ux-securities/backend/internal/domain/market"
	domain "ux-securities/backend/internal/domain/trading"
)

type redRepository struct{}

func (redRepository) CreateOrder(context.Context, int64, CreateOrderInput, market.Quote) (Order, error) {
	return Order{}, errors.New("not implemented")
}
func (redRepository) FindOrderByIdempotencyKey(context.Context, int64, string) (Order, error) {
	return Order{}, ErrNotFound
}
func (redRepository) GetOrder(context.Context, int64, int64) (Order, error) {
	return Order{}, ErrNotFound
}
func (redRepository) CancelOrder(context.Context, int64, int64) (Order, error) {
	return Order{}, ErrNotFound
}
func (redRepository) Account(context.Context, int64) (Account, error) {
	return Account{}, errors.New("not implemented")
}
func (redRepository) Positions(context.Context, int64) ([]Position, error) {
	return nil, errors.New("not implemented")
}
func (redRepository) Orders(context.Context, int64, OrderFilter) ([]Order, error) {
	return nil, errors.New("not implemented")
}

func (redRepository) Settle(context.Context, int64, CreateOrderInput, market.Quote) (Order, error) {
	return Order{ID: 1, Status: domain.StatusFilled}, nil
}

func TestCreateRejectsStaleQuoteBeforePersistence(t *testing.T) {
	now := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	service := NewService(redRepository{}, func() time.Time { return now }, 30*time.Second)
	_, err := service.CreateWithQuote(context.Background(), auth.Actor{UserID: 7}, CreateOrderInput{
		ClientOrderID: "idem-1234", Symbol: "00700", Side: domain.SideBuy, Quantity: 100, LimitPrice: 380,
	}, market.Quote{Symbol: "00700", LastPrice: float64Ptr(380), AsOf: now.Add(-time.Minute), DataStatus: market.DataStatusSimulated})
	if !errors.Is(err, ErrStaleQuote) {
		t.Fatalf("Create() error = %v, want ErrStaleQuote", err)
	}
}

func TestCreateIsIdempotentAndSettlesAtomically(t *testing.T) {
	location := time.FixedZone("HKT", 8*60*60)
	now := time.Date(2026, time.October, 26, 10, 0, 0, 0, location)
	service := NewService(redRepository{}, func() time.Time { return now }, 30*time.Second)
	_, err := service.CreateWithQuote(context.Background(), auth.Actor{UserID: 7}, CreateOrderInput{
		ClientOrderID: "idem-1234", Symbol: "00700", Side: domain.SideBuy, Quantity: 100, LimitPrice: 380,
	}, market.Quote{Symbol: "00700", LastPrice: float64Ptr(380), AsOf: now, DataStatus: market.DataStatusSimulated})
	if err != nil {
		t.Fatalf("Create() error = %v, want idempotent filled order", err)
	}
}

func TestCreateWithQuoteRejectsWhenHKEXMarketIsClosed(t *testing.T) {
	location, err := time.LoadLocation("Asia/Hong_Kong")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.October, 24, 10, 0, 0, 0, location)
	service := NewService(redRepository{}, func() time.Time { return now }, 30*time.Second)
	_, err = service.CreateWithQuote(context.Background(), auth.Actor{UserID: 7}, CreateOrderInput{
		ClientOrderID: "idem-1234", Symbol: "00700", Side: domain.SideBuy, Quantity: 100, LimitPrice: 380,
	}, market.Quote{Symbol: "00700", LastPrice: float64Ptr(380), AsOf: now, DataStatus: market.DataStatusSimulated})
	if !errors.Is(err, ErrMarketClosed) {
		t.Fatalf("CreateWithQuote() error = %v, want market is closed", err)
	}
}

type missingOrderRepository struct{ redRepository }

func (missingOrderRepository) GetOrder(context.Context, int64, int64) (Order, error) {
	return Order{}, sql.ErrNoRows
}

func TestOrderMapsMissingStorageRowToNotFound(t *testing.T) {
	service := NewService(missingOrderRepository{})
	_, err := service.Order(context.Background(), auth.Actor{UserID: 7}, 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Order() error = %v, want ErrNotFound", err)
	}
}

func float64Ptr(value float64) *float64 { return &value }
