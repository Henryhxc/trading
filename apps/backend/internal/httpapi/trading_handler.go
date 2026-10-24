package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
	auth "ux-securities/backend/internal/domain/auth"
	domain "ux-securities/backend/internal/domain/trading"
	"ux-securities/backend/internal/service/trading"
)

type TradingAPI interface {
	Create(context.Context, auth.Actor, trading.CreateOrderInput) (trading.Order, error)
	Cancel(context.Context, auth.Actor, int64) (trading.Order, error)
	Order(context.Context, auth.Actor, int64) (trading.Order, error)
	Account(context.Context, auth.Actor) (trading.Account, error)
	Positions(context.Context, auth.Actor) ([]trading.Position, error)
	Orders(context.Context, auth.Actor, trading.OrderFilter) ([]trading.Order, error)
}

func (s *server) createOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClientOrderID string      `json:"client_order_id"`
		Symbol        string      `json:"symbol"`
		Side          domain.Side `json:"side"`
		OrderType     string      `json:"order_type"`
		Quantity      int64       `json:"quantity"`
		LimitPrice    float64     `json:"limit_price"`
	}
	if !decode(w, r, &body) {
		s.fail(w, r, http.StatusBadRequest, "invalid_request", "Request body is invalid")
		return
	}
	if body.OrderType != "limit" {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Only limit orders are supported")
		return
	}
	clientOrderID := r.Header.Get("Idempotency-Key")
	if clientOrderID == "" {
		clientOrderID = body.ClientOrderID
	} else if body.ClientOrderID != "" && body.ClientOrderID != clientOrderID {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Idempotency key does not match request body")
		return
	}
	order, err := s.trading.Create(r.Context(), actorFrom(r), trading.CreateOrderInput{ClientOrderID: clientOrderID, Symbol: body.Symbol, Side: body.Side, Quantity: body.Quantity, LimitPrice: body.LimitPrice})
	if err != nil {
		s.tradingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (s *server) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Check the submitted details")
		return
	}
	order, err := s.trading.Order(r.Context(), actorFrom(r), id)
	if err != nil {
		s.tradingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *server) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Check the submitted details")
		return
	}
	order, err := s.trading.Cancel(r.Context(), actorFrom(r), id)
	if err != nil {
		s.tradingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *server) account(w http.ResponseWriter, r *http.Request) {
	value, err := s.trading.Account(r.Context(), actorFrom(r))
	if err != nil {
		s.tradingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
func (s *server) positions(w http.ResponseWriter, r *http.Request) {
	value, err := s.trading.Positions(r.Context(), actorFrom(r))
	if err != nil {
		s.tradingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": value})
}

func (s *server) orders(w http.ResponseWriter, r *http.Request) {
	filter := trading.OrderFilter{Symbol: r.URL.Query().Get("symbol"), Side: domain.Side(r.URL.Query().Get("side")), Status: domain.OrderStatus(r.URL.Query().Get("status"))}
	var err error
	if raw := r.URL.Query().Get("from"); raw != "" {
		filter.From, err = time.Parse(time.RFC3339, raw)
	}
	if err == nil {
		if raw := r.URL.Query().Get("to"); raw != "" {
			filter.To, err = time.Parse(time.RFC3339, raw)
		}
	}
	if err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Check the submitted details")
		return
	}
	value, err := s.trading.Orders(r.Context(), actorFrom(r), filter)
	if err != nil {
		s.tradingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": value})
}

func (s *server) tradingError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusInternalServerError, "internal_error", "Request could not be completed"
	switch {
	case errors.Is(err, trading.ErrNotFound):
		status, code, message = http.StatusNotFound, "not_found", "Resource was not found"
	case errors.Is(err, trading.ErrInvalidInput):
		status, code, message = http.StatusBadRequest, "invalid_input", "Check the submitted details"
	case errors.Is(err, trading.ErrStaleQuote):
		status, code, message = http.StatusUnprocessableEntity, "stale_quote", "Market data is stale"
	case errors.Is(err, trading.ErrMarketClosed):
		status, code, message = http.StatusUnprocessableEntity, "market_closed", "HKEX continuous trading is closed"
	case errors.Is(err, trading.ErrInsufficientFunds), errors.Is(err, trading.ErrInsufficientPosition):
		status, code, message = http.StatusUnprocessableEntity, "insufficient_balance", "Order could not be filled"
	case errors.Is(err, trading.ErrInvalidState):
		status, code, message = http.StatusConflict, "invalid_state", "Order cannot be changed"
	}
	s.fail(w, r, status, code, message)
}
