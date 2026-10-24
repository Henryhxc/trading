package mysqlrepo

import (
	"context"
	"database/sql"
	"time"
	domain "ux-securities/backend/internal/domain/cashmovement"
)

type CashMovementRepository struct{ db *sql.DB }

func NewCashMovementRepository(db *sql.DB) *CashMovementRepository { return &CashMovementRepository{db: db} }

func (r *CashMovementRepository) ListCashMovements(ctx context.Context, userID int64) ([]domain.Movement, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return nil, err }
	defer tx.Rollback()
	now := time.Now().UTC()
	seed := []struct { kind domain.Type; amount float64; method string; status domain.Status }{
		{domain.TypeDeposit, 10000, "Bank transfer (mock)", domain.StatusCompleted},
		{domain.TypeWithdraw, 1200, "FPS beneficiary (mock)", domain.StatusCompleted},
		{domain.TypeDeposit, 2500, "FPS (mock)", domain.StatusPending},
	}
	for _, item := range seed {
		_, err = tx.ExecContext(ctx, "INSERT INTO cash_movements (user_id,movement_type,amount,currency,method,status,prototype,created_at) SELECT ?,?,?, 'HKD', ?,?,1,? WHERE NOT EXISTS (SELECT 1 FROM cash_movements WHERE user_id=? AND movement_type=? AND method=?)", userID, item.kind, item.amount, item.method, item.status, now, userID, item.kind, item.method)
		if err != nil { return nil, err }
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,user_id,movement_type,amount,currency,method,status,prototype,created_at FROM cash_movements WHERE user_id=? ORDER BY created_at DESC, id DESC", userID)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []domain.Movement{}
	for rows.Next() {
		var item domain.Movement
		var kind, status string
		var prototype bool
		if err := rows.Scan(&item.ID, &item.UserID, &kind, &item.Amount, &item.Currency, &item.Method, &status, &prototype, &item.CreatedAt); err != nil { return nil, err }
		item.Type, item.Status, item.Prototype = domain.Type(kind), domain.Status(status), prototype
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, err }
	if err := tx.Commit(); err != nil { return nil, err }
	return items, nil
}
