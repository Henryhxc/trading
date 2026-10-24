package mysqlrepo

import (
	"context"
	"database/sql"
	"strings"
	"time"

	domain "ux-securities/backend/internal/domain/auth"
)

type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) PingContext(ctx context.Context) error { return r.db.PingContext(ctx) }

func (r *Repository) CreateUser(ctx context.Context, input domain.Registration, passwordHash, tokenHash, initialCash string, expires time.Time) (domain.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "INSERT INTO users (username, email, password_hash, status, created_at, updated_at) VALUES (?, ?, ?, 'active', UTC_TIMESTAMP(3), UTC_TIMESTAMP(3))", input.Username, input.Email, passwordHash)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return domain.User{}, domain.ErrIdentityConflict
		}
		return domain.User{}, err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return domain.User{}, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO accounts (user_id, initial_cash, cash_balance, available_cash, currency, status, updated_at) VALUES (?, ?, ?, ?, 'HKD', 'active', UTC_TIMESTAMP(3))", userID, initialCash, initialCash, initialCash); err != nil {
		return domain.User{}, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO sessions (user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, UTC_TIMESTAMP(3))", userID, tokenHash, expires); err != nil {
		return domain.User{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.User{}, err
	}
	return domain.User{ID: userID, Username: input.Username, Email: input.Email}, nil
}

func (r *Repository) FindUserByIdentity(ctx context.Context, identity string) (domain.User, string, error) {
	var user domain.User
	var passwordHash string
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email, password_hash FROM users WHERE username = ? OR email = ? LIMIT 1", identity, strings.ToLower(identity)).Scan(&user.ID, &user.Username, &user.Email, &passwordHash)
	return user, passwordHash, err
}

func (r *Repository) CreateSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO sessions (user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, UTC_TIMESTAMP(3))", userID, tokenHash, expires)
	return err
}

func (r *Repository) ResolveSession(ctx context.Context, tokenHash string, now time.Time) (domain.Actor, error) {
	var actor domain.Actor
	err := r.db.QueryRowContext(ctx, "SELECT u.id, u.username, u.email FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash = ? AND s.revoked_at IS NULL AND s.expires_at > ? AND u.status = 'active' LIMIT 1", tokenHash, now).Scan(&actor.UserID, &actor.Username, &actor.Email)
	return actor, err
}

func (r *Repository) RevokeSession(ctx context.Context, tokenHash string, now time.Time) error {
	_, err := r.db.ExecContext(ctx, "UPDATE sessions SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL", now, tokenHash)
	return err
}
