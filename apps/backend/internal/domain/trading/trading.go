package trading

import "time"

type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusFilled    OrderStatus = "filled"
	StatusCancelled OrderStatus = "cancelled"
	StatusRejected  OrderStatus = "rejected"
)

type Order struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"-"`
	Symbol          string      `json:"symbol"`
	Name            string      `json:"name"`
	Side            Side        `json:"side"`
	Quantity        int64       `json:"quantity"`
	LimitPrice      float64     `json:"limit_price"`
	EstimatedAmount float64     `json:"estimated_amount"`
	Status          OrderStatus `json:"status"`
	ClientOrderID   string      `json:"client_order_id"`
	FilledQuantity  int64       `json:"filled_quantity"`
	FilledPrice     *float64    `json:"filled_price"`
	FilledAt        *time.Time  `json:"filled_at,omitempty"`
	RejectReason    string      `json:"reject_reason,omitempty"`
	QuoteAsOf       time.Time   `json:"quote_as_of"`
	CreatedAt       time.Time   `json:"created_at"`
}
type Account struct {
	ID            int64   `json:"id"`
	Currency      string  `json:"currency"`
	CashBalance   float64 `json:"cash_balance"`
	AvailableCash float64 `json:"available_cash"`
	InitialCash   float64 `json:"initial_cash"`
	Status        string  `json:"status"`
}
type Position struct {
	Symbol            string   `json:"symbol"`
	Name              string   `json:"display_name"`
	Quantity          int64    `json:"quantity"`
	AvailableQuantity int64    `json:"available_quantity"`
	AverageCost       float64  `json:"average_cost"`
	CurrentPrice      *float64 `json:"market_price"`
	MarketValue       *float64 `json:"market_value"`
	UnrealizedPnL     *float64 `json:"unrealized_pnl"`
	DataStatus        string   `json:"data_status"`
}
type Filter struct {
	Symbol   string
	Side     Side
	Status   OrderStatus
	From, To time.Time
}
