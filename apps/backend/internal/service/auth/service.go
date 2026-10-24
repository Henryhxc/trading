package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	domain "ux-securities/backend/internal/domain/auth"
)

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrCredentials     = errors.New("invalid credentials")
	ErrConflict        = domain.ErrIdentityConflict
	ErrUnauthenticated = errors.New("unauthenticated")
)

const (
	minPasswordBytes = 12
	maxPasswordBytes = 72
)

type Repository interface {
	CreateUser(context.Context, domain.Registration, string, string, string, time.Time) (domain.User, error)
	FindUserByIdentity(context.Context, string) (domain.User, string, error)
	CreateSession(context.Context, int64, string, time.Time) error
	ResolveSession(context.Context, string, time.Time) (domain.Actor, error)
	RevokeSession(context.Context, string, time.Time) error
}

type Service struct {
	repo        Repository
	ttl         time.Duration
	initialCash string
}

func NewService(repo Repository, ttl time.Duration, initialCash string) *Service {
	return &Service{repo: repo, ttl: ttl, initialCash: initialCash}
}

func (s *Service) Register(ctx context.Context, input domain.Registration) (domain.User, string, time.Time, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if len(input.Username) < 3 || len(input.Username) > 64 || !validEmail(input.Email) || len(input.Password) < minPasswordBytes || len(input.Password) > maxPasswordBytes {
		return domain.User{}, "", time.Time{}, ErrInvalidInput
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}
	token, hash, expires := newSession(s.ttl)
	user, err := s.repo.CreateUser(ctx, input, string(passwordHash), hash, s.initialCash, expires)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}
	return user, token, expires, nil
}

func validEmail(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value
}

func (s *Service) Login(ctx context.Context, input domain.Credentials) (domain.User, string, time.Time, error) {
	user, passwordHash, err := s.repo.FindUserByIdentity(ctx, strings.TrimSpace(input.Identity))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, "", time.Time{}, ErrCredentials
		}
		return domain.User{}, "", time.Time{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		return domain.User{}, "", time.Time{}, ErrCredentials
	}
	token, hash, expires := newSession(s.ttl)
	if err := s.repo.CreateSession(ctx, user.ID, hash, expires); err != nil {
		return domain.User{}, "", time.Time{}, err
	}
	return user, token, expires, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (domain.Actor, error) {
	if token == "" {
		return domain.Actor{}, ErrUnauthenticated
	}
	actor, err := s.repo.ResolveSession(ctx, tokenHash(token), time.Now().UTC())
	if err != nil {
		return domain.Actor{}, ErrUnauthenticated
	}
	return actor, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.repo.RevokeSession(ctx, tokenHash(token), time.Now().UTC())
}

func (s *Service) Health(ctx context.Context) error {
	checker, ok := s.repo.(interface{ PingContext(context.Context) error })
	if !ok {
		return errors.New("repository readiness check unavailable")
	}
	return checker.PingContext(ctx)
}

func newSession(ttl time.Duration) (string, string, time.Time) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		panic("secure random source unavailable")
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, tokenHash(token), time.Now().UTC().Add(ttl)
}

func tokenHash(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
