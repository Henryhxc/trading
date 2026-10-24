package cashmovement

import (
	"context"
	auth "ux-securities/backend/internal/domain/auth"
	domain "ux-securities/backend/internal/domain/cashmovement"
)

type Repository interface {
	ListCashMovements(context.Context, int64) ([]domain.Movement, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, actor auth.Actor) ([]domain.Movement, error) {
	return s.repo.ListCashMovements(ctx, actor.UserID)
}
