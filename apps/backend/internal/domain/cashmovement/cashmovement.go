package cashmovement

import "time"

type Type string

const (
	TypeDeposit  Type = "deposit"
	TypeWithdraw Type = "withdraw"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
)

type Movement struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	Type      Type      `json:"type"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Method    string    `json:"method"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Prototype bool      `json:"prototype"`
}
