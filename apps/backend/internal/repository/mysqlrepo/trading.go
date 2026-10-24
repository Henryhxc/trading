package mysqlrepo

import (
	"context"
	"database/sql"
	"errors"
	market "ux-securities/backend/internal/domain/market"
	domain "ux-securities/backend/internal/domain/trading"
	trading "ux-securities/backend/internal/service/trading"
)

type TradingRepository struct{ db *sql.DB }

func NewTradingRepository(db *sql.DB) *TradingRepository { return &TradingRepository{db: db} }
func (r *TradingRepository) FindOrderByIdempotencyKey(ctx context.Context, userID int64, key string) (trading.Order, error) {
	return r.order(ctx, "WHERE user_id = ? AND client_order_id = ?", userID, key)
}
func (r *TradingRepository) GetOrder(ctx context.Context, userID, id int64) (trading.Order, error) {
	return r.order(ctx, "WHERE o.user_id = ? AND o.id = ?", userID, id)
}
func (r *TradingRepository) CreateOrder(ctx context.Context, userID int64, input trading.CreateOrderInput, quote market.Quote) (trading.Order, error) {
	return r.insertOrder(ctx, nil, userID, input, quote, domain.StatusPending, "")
}
func (r *TradingRepository) Settle(ctx context.Context, userID int64, input trading.CreateOrderInput, quote market.Quote) (trading.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return trading.Order{}, err
	}
	defer tx.Rollback()
	if existing, err := r.orderTx(ctx, tx, "WHERE o.user_id = ? AND o.client_order_id = ?", userID, input.ClientOrderID); err == nil {
		return existing, nil
	}
	var lot int64
	var name string
	var last float64
	if err = tx.QueryRowContext(ctx, "SELECT name, lot_size, last_price FROM stocks WHERE symbol = ? AND status = 'active'", input.Symbol).Scan(&name, &lot, &last); errors.Is(err, sql.ErrNoRows) {
		return trading.Order{}, trading.ErrNotFound
	} else if err != nil {
		return trading.Order{}, err
	}
	if input.Quantity%lot != 0 {
		return trading.Order{}, trading.ErrInvalidInput
	}
	matched := (input.Side == domain.SideBuy && input.LimitPrice >= last) || (input.Side == domain.SideSell && input.LimitPrice <= last)
	status := domain.StatusPending
	filled := int64(0)
	var filledPrice any
	if matched {
		status, filled, filledPrice = domain.StatusFilled, input.Quantity, last
	}
	order, err := r.insertOrder(ctx, tx, userID, input, quote, status, name)
	if err != nil {
		return trading.Order{}, err
	}
	if matched {
		if err = r.applyFill(ctx, tx, userID, input, last, order.ID); err != nil {
			return trading.Order{}, err
		}
		order.Status, order.FilledQuantity, order.FilledPrice = status, filled, &last
		_ = filledPrice
	}
	if err = tx.Commit(); err != nil {
		return trading.Order{}, err
	}
	return order, nil
}
func (r *TradingRepository) insertOrder(ctx context.Context, tx *sql.Tx, userID int64, input trading.CreateOrderInput, quote market.Quote, status domain.OrderStatus, name string) (trading.Order, error) {
	var exec interface {
		ExecContext(context.Context, string, ...any) (sql.Result, error)
	} = r.db
	if tx != nil {
		exec = tx
	}
	var filledPrice any
	if status == domain.StatusFilled && quote.LastPrice != nil {
		filledPrice = *quote.LastPrice
	}
	result, err := exec.ExecContext(ctx, "INSERT INTO orders (user_id,symbol,side,quantity,limit_price,estimated_amount,status,client_order_id,filled_quantity,filled_price,filled_at,quote_as_of,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,IF(?='filled',UTC_TIMESTAMP(3),NULL),?,UTC_TIMESTAMP(3),UTC_TIMESTAMP(3))", userID, input.Symbol, input.Side, input.Quantity, input.LimitPrice, input.LimitPrice*float64(input.Quantity), status, input.ClientOrderID, map[bool]int64{true: input.Quantity, false: 0}[status == domain.StatusFilled], filledPrice, status, quote.AsOf)
	if err != nil {
		return trading.Order{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return trading.Order{}, err
	}
	return trading.Order{ID: id, UserID: userID, Symbol: input.Symbol, Name: name, Side: input.Side, Quantity: input.Quantity, LimitPrice: input.LimitPrice, EstimatedAmount: input.LimitPrice * float64(input.Quantity), Status: status, ClientOrderID: input.ClientOrderID, QuoteAsOf: quote.AsOf}, nil
}
func (r *TradingRepository) applyFill(ctx context.Context, tx *sql.Tx, userID int64, input trading.CreateOrderInput, price float64, orderID int64) error {
	amount := price * float64(input.Quantity)
	if input.Side == domain.SideBuy {
		result, err := tx.ExecContext(ctx, "UPDATE accounts SET cash_balance=cash_balance-?, available_cash=available_cash-?, updated_at=UTC_TIMESTAMP(3) WHERE user_id=? AND available_cash>=? AND status='active'", amount, amount, userID, amount)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			return trading.ErrInsufficientFunds
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO positions (user_id,symbol,quantity,available_quantity,average_cost,updated_at) VALUES (?,?,?,?,?,UTC_TIMESTAMP(3)) ON DUPLICATE KEY UPDATE average_cost=((quantity*average_cost)+(?*?))/(quantity+?), quantity=quantity+?,available_quantity=available_quantity+?,updated_at=UTC_TIMESTAMP(3)", userID, input.Symbol, input.Quantity, input.Quantity, price, input.Quantity, price, input.Quantity, input.Quantity, input.Quantity)
		return err
	}
	result, err := tx.ExecContext(ctx, "UPDATE positions SET quantity=quantity-?,available_quantity=available_quantity-?,updated_at=UTC_TIMESTAMP(3) WHERE user_id=? AND symbol=? AND available_quantity>=?", input.Quantity, input.Quantity, userID, input.Symbol, input.Quantity)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return trading.ErrInsufficientPosition
	}
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET cash_balance=cash_balance+?,available_cash=available_cash+?,updated_at=UTC_TIMESTAMP(3) WHERE user_id=?", amount, amount, userID)
	_ = orderID
	return err
}
func (r *TradingRepository) CancelOrder(ctx context.Context, userID, id int64) (trading.Order, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE orders SET status='cancelled',updated_at=UTC_TIMESTAMP(3) WHERE id=? AND user_id=? AND status='pending'", id, userID)
	if err != nil {
		return trading.Order{}, err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return trading.Order{}, trading.ErrInvalidState
	}
	return r.GetOrder(ctx, userID, id)
}
func (r *TradingRepository) Account(ctx context.Context, userID int64) (trading.Account, error) {
	var a trading.Account
	err := r.db.QueryRowContext(ctx, "SELECT id,currency,cash_balance,available_cash,initial_cash,status FROM accounts WHERE user_id=?", userID).Scan(&a.ID, &a.Currency, &a.CashBalance, &a.AvailableCash, &a.InitialCash, &a.Status)
	return a, err
}
func (r *TradingRepository) Positions(ctx context.Context, userID int64) ([]trading.Position, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT p.symbol,s.name,p.quantity,p.available_quantity,p.average_cost,s.last_price,s.data_status FROM positions p JOIN stocks s ON s.symbol=p.symbol WHERE p.user_id=? AND p.quantity>0 ORDER BY p.symbol", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []trading.Position{}
	for rows.Next() {
		var p trading.Position
		var price sql.NullFloat64
		if err := rows.Scan(&p.Symbol, &p.Name, &p.Quantity, &p.AvailableQuantity, &p.AverageCost, &price, &p.DataStatus); err != nil {
			return nil, err
		}
		if price.Valid {
			p.CurrentPrice = &price.Float64
			value := price.Float64 * float64(p.Quantity)
			pnl := (price.Float64 - p.AverageCost) * float64(p.Quantity)
			p.MarketValue = &value
			p.UnrealizedPnL = &pnl
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
func (r *TradingRepository) Orders(ctx context.Context, userID int64, f domain.Filter) ([]trading.Order, error) {
	query := "WHERE o.user_id=?"
	args := []any{userID}
	if f.Symbol != "" {
		query += " AND o.symbol=?"
		args = append(args, f.Symbol)
	}
	if f.Side != "" {
		query += " AND o.side=?"
		args = append(args, f.Side)
	}
	if f.Status != "" {
		query += " AND o.status=?"
		args = append(args, f.Status)
	}
	if !f.From.IsZero() {
		query += " AND o.created_at>=?"
		args = append(args, f.From)
	}
	if !f.To.IsZero() {
		query += " AND o.created_at<=?"
		args = append(args, f.To)
	}
	query += " ORDER BY o.created_at DESC"
	return r.orders(ctx, query, args...)
}
func (r *TradingRepository) order(ctx context.Context, where string, args ...any) (trading.Order, error) {
	return r.orderScan(r.db.QueryRowContext(ctx, "SELECT o.id,o.user_id,o.symbol,COALESCE(s.name,''),o.side,o.quantity,o.limit_price,o.estimated_amount,o.status,o.client_order_id,o.filled_quantity,o.filled_price,o.filled_at,o.reject_reason,o.quote_as_of,o.created_at FROM orders o LEFT JOIN stocks s ON s.symbol=o.symbol "+where, args...))
}
func (r *TradingRepository) orderTx(ctx context.Context, tx *sql.Tx, where string, args ...any) (trading.Order, error) {
	return r.orderScan(tx.QueryRowContext(ctx, "SELECT o.id,o.user_id,o.symbol,COALESCE(s.name,''),o.side,o.quantity,o.limit_price,o.estimated_amount,o.status,o.client_order_id,o.filled_quantity,o.filled_price,o.filled_at,o.reject_reason,o.quote_as_of,o.created_at FROM orders o LEFT JOIN stocks s ON s.symbol=o.symbol "+where+" FOR UPDATE", args...))
}
func (r *TradingRepository) orderScan(row *sql.Row) (trading.Order, error) {
	var o trading.Order
	var price sql.NullFloat64
	var filledAt sql.NullTime
	var reason sql.NullString
	err := row.Scan(&o.ID, &o.UserID, &o.Symbol, &o.Name, &o.Side, &o.Quantity, &o.LimitPrice, &o.EstimatedAmount, &o.Status, &o.ClientOrderID, &o.FilledQuantity, &price, &filledAt, &reason, &o.QuoteAsOf, &o.CreatedAt)
	if price.Valid {
		o.FilledPrice = &price.Float64
	}
	if filledAt.Valid {
		o.FilledAt = &filledAt.Time
	}
	o.RejectReason = reason.String
	return o, err
}
func (r *TradingRepository) orders(ctx context.Context, where string, args ...any) ([]trading.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT o.id,o.user_id,o.symbol,COALESCE(s.name,''),o.side,o.quantity,o.limit_price,o.estimated_amount,o.status,o.client_order_id,o.filled_quantity,o.filled_price,o.filled_at,o.reject_reason,o.quote_as_of,o.created_at FROM orders o LEFT JOIN stocks s ON s.symbol=o.symbol "+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []trading.Order{}
	for rows.Next() {
		var o trading.Order
		var price sql.NullFloat64
		var at sql.NullTime
		var reason sql.NullString
		if err := rows.Scan(&o.ID, &o.UserID, &o.Symbol, &o.Name, &o.Side, &o.Quantity, &o.LimitPrice, &o.EstimatedAmount, &o.Status, &o.ClientOrderID, &o.FilledQuantity, &price, &at, &reason, &o.QuoteAsOf, &o.CreatedAt); err != nil {
			return nil, err
		}
		if price.Valid {
			o.FilledPrice = &price.Float64
		}
		if at.Valid {
			o.FilledAt = &at.Time
		}
		o.RejectReason = reason.String
		result = append(result, o)
	}
	return result, rows.Err()
}
