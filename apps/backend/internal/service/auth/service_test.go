package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/auth"
)

type registrationRecorder struct{ input domain.Registration }

func (r *registrationRecorder) CreateUser(_ context.Context, input domain.Registration, _, _, _ string, _ time.Time) (domain.User, error) {
	r.input = input
	return domain.User{ID: 1, Username: input.Username, Email: input.Email}, nil
}
func (*registrationRecorder) FindUserByIdentity(context.Context, string) (domain.User, string, error) {
	return domain.User{}, "", nil
}
func (*registrationRecorder) CreateSession(context.Context, int64, string, time.Time) error {
	return nil
}
func (*registrationRecorder) ResolveSession(context.Context, string, time.Time) (domain.Actor, error) {
	return domain.Actor{}, nil
}
func (*registrationRecorder) RevokeSession(context.Context, string, time.Time) error { return nil }

type loginRepository struct {
	findErr      error
	passwordHash string
}

func (*loginRepository) CreateUser(context.Context, domain.Registration, string, string, string, time.Time) (domain.User, error) {
	return domain.User{}, nil
}
func (r *loginRepository) FindUserByIdentity(context.Context, string) (domain.User, string, error) {
	return domain.User{ID: 1, Username: "trader", Email: "trader@example.com"}, r.passwordHash, r.findErr
}
func (*loginRepository) CreateSession(context.Context, int64, string, time.Time) error { return nil }
func (*loginRepository) ResolveSession(context.Context, string, time.Time) (domain.Actor, error) {
	return domain.Actor{}, nil
}
func (*loginRepository) RevokeSession(context.Context, string, time.Time) error { return nil }

func TestLoginMapsMissingIdentityToInvalidCredentials(t *testing.T) {
	service := NewService(&loginRepository{findErr: sql.ErrNoRows}, time.Hour, "1000000.0000")
	_, _, _, err := service.Login(context.Background(), domain.Credentials{Identity: "missing", Password: "long-enough-pass"})
	if !errors.Is(err, ErrCredentials) {
		t.Fatalf("Login() error = %v, want ErrCredentials for missing identity", err)
	}
}

func TestLoginPreservesUnexpectedRepositoryError(t *testing.T) {
	storageErr := errors.New("database connection reset")
	service := NewService(&loginRepository{findErr: storageErr}, time.Hour, "1000000.0000")
	_, _, _, err := service.Login(context.Background(), domain.Credentials{Identity: "trader", Password: "long-enough-pass"})
	if !errors.Is(err, storageErr) {
		t.Fatalf("Login() error = %v, want repository error %v", err, storageErr)
	}
}

func TestRegisterRejectsMalformedEmailBeforePersisting(t *testing.T) {
	invalid := []string{"x@", "@example.com", "person..name@example.com", "person@.example.com", "Name <person@example.com>"}
	for _, email := range invalid {
		t.Run(email, func(t *testing.T) {
			repo := &registrationRecorder{}
			service := NewService(repo, time.Hour, "1000000.0000")
			_, _, _, err := service.Register(context.Background(), domain.Registration{
				Username: "valid-user", Email: email, Password: "long-enough-pass",
			})
			if err != ErrInvalidInput {
				t.Fatalf("Register() error = %v, want ErrInvalidInput", err)
			}
			if repo.input.Email != "" {
				t.Fatalf("invalid email reached repository: %q", repo.input.Email)
			}
		})
	}
}

func TestRegisterAcceptsEmailAtDatabaseLengthLimit(t *testing.T) {
	localPart := "person"
	for len(localPart)+len("@example.com") < 128 {
		localPart += "a"
	}
	email := localPart + "@example.com"
	repo := &registrationRecorder{}
	service := NewService(repo, time.Hour, "1000000.0000")
	_, _, _, err := service.Register(context.Background(), domain.Registration{
		Username: "valid-user", Email: email, Password: "long-enough-pass",
	})
	if err != nil {
		t.Fatalf("Register() rejected valid %d-byte email: %v", len(email), err)
	}
	if repo.input.Email != email {
		t.Fatalf("repository email = %q, want normalized %q", repo.input.Email, email)
	}
}
