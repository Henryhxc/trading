package trading

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	auth "ux-securities/backend/internal/domain/auth"
	market "ux-securities/backend/internal/domain/market"
	domain "ux-securities/backend/internal/domain/trading"
)

var (
	ErrNotFound             = errors.New("trading resource not found")
	ErrInvalidInput         = errors.New("invalid trading input")
	ErrStaleQuote           = errors.New("quote is stale")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInsufficientPosition = errors.New("insufficient position")
	ErrInvalidState         = errors.New("invalid order state")
	ErrMarketClosed         = errors.New("market is closed")
)

type CreateOrderInput struct {
	ClientOrderID string
	Symbol        string
	Side          domain.Side
	Quantity      int64
	LimitPrice    float64
}
type Order = domain.Order
type Account = domain.Account
type Position = domain.Position
type OrderFilter = domain.Filter
type Side = domain.Side
type Repository interface {
	CreateOrder(context.Context, int64, CreateOrderInput, market.Quote) (Order, error)
	FindOrderByIdempotencyKey(context.Context, int64, string) (Order, error)
	GetOrder(context.Context, int64, int64) (Order, error)
	CancelOrder(context.Context, int64, int64) (Order, error)
	Account(context.Context, int64) (Account, error)
	Positions(context.Context, int64) ([]Position, error)
	Orders(context.Context, int64, OrderFilter) ([]Order, error)
	Settle(context.Context, int64, CreateOrderInput, market.Quote) (Order, error)
}
type Market interface {
	Quote(context.Context, string) (market.Quote, error)
}
type Service struct {
	repo       Repository
	market     Market
	now        func() time.Time
	staleAfter time.Duration
}

func NewService(repo Repository, args ...any) *Service {
	s := &Service{repo: repo, now: time.Now, staleAfter: 30 * time.Second}
	for _, arg := range args {
		switch v := arg.(type) {
		case Market:
			s.market = v
		case func() time.Time:
			s.now = v
		case time.Duration:
			s.staleAfter = v
		}
	}
	return s
}
func (s *Service) Create(ctx context.Context, actor auth.Actor, input CreateOrderInput) (Order, error) {
	if err := validate(input); err != nil {
		return Order{}, err
	}
	input.Symbol = normalize(input.Symbol)
	if existing, err := s.repo.FindOrderByIdempotencyKey(ctx, actor.UserID, input.ClientOrderID); err == nil {
		return existing, nil
	}
	if s.market == nil {
		return Order{}, ErrInvalidInput
	}
	quote, err := s.market.Quote(ctx, input.Symbol)
	if err != nil {
		return Order{}, err
	}
	if quote.AsOf.IsZero() || s.now().Sub(quote.AsOf) > s.staleAfter || quote.DataStatus == market.DataStatusStale || quote.DataStatus == market.DataStatusUnavailable {
		return Order{}, ErrStaleQuote
	}
	if !isHKContinuousTradingSession(s.now()) {
		return Order{}, ErrMarketClosed
	}
	if quote.LastPrice == nil || input.LimitPrice <= 0 {
		return Order{}, ErrInvalidInput
	}
	return s.repo.Settle(ctx, actor.UserID, input, quote)
}
func (s *Service) CreateWithQuote(ctx context.Context, actor auth.Actor, input CreateOrderInput, quote market.Quote) (Order, error) {
	if err := validate(input); err != nil {
		return Order{}, err
	}
	if quote.AsOf.IsZero() || s.now().Sub(quote.AsOf) > s.staleAfter {
		return Order{}, ErrStaleQuote
	}
	if !isHKContinuousTradingSession(s.now()) {
		return Order{}, ErrMarketClosed
	}
	if existing, err := s.repo.FindOrderByIdempotencyKey(ctx, actor.UserID, input.ClientOrderID); err == nil {
		return existing, nil
	}
	return s.repo.Settle(ctx, actor.UserID, input, quote)
}
func (s *Service) Cancel(ctx context.Context, actor auth.Actor, id int64) (Order, error) {
	return s.repo.CancelOrder(ctx, actor.UserID, id)
}
func (s *Service) Order(ctx context.Context, actor auth.Actor, id int64) (Order, error) {
	order, err := s.repo.GetOrder(ctx, actor.UserID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	return order, err
}
func (s *Service) Account(ctx context.Context, actor auth.Actor) (Account, error) {
	return s.repo.Account(ctx, actor.UserID)
}
func (s *Service) Positions(ctx context.Context, actor auth.Actor) ([]Position, error) {
	return s.repo.Positions(ctx, actor.UserID)
}
func (s *Service) Orders(ctx context.Context, actor auth.Actor, filter OrderFilter) ([]Order, error) {
	return s.repo.Orders(ctx, actor.UserID, filter)
}
func isHKContinuousTradingSession(now time.Time) bool {
	hongKong := now.In(time.FixedZone("HKT", 8*60*60))
	weekday := hongKong.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	minute := hongKong.Hour()*60 + hongKong.Minute()
	return (minute >= 9*60+30 && minute < 12*60) || (minute >= 13*60 && minute < 16*60)
}
func validate(input CreateOrderInput) error {
	if len(strings.TrimSpace(input.ClientOrderID)) < 8 || len(input.ClientOrderID) > 128 || input.Quantity <= 0 || input.LimitPrice <= 0 || (input.Side != domain.SideBuy && input.Side != domain.SideSell) {
		return ErrInvalidInput
	}
	return nil
}
func normalize(symbol string) string {
	symbol = strings.TrimSpace(symbol)
	for len(symbol) < 5 {
		symbol = "0" + symbol
	}
	return symbol
}
