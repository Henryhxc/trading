package cashmovement

import (
	"context"
	"testing"
	auth "ux-securities/backend/internal/domain/auth"
	domain "ux-securities/backend/internal/domain/cashmovement"
)

type repositoryStub struct { userID int64; items []domain.Movement }

func (r *repositoryStub) ListCashMovements(_ context.Context, userID int64) ([]domain.Movement, error) {
	r.userID = userID
	return r.items, nil
}

func TestListLoadsDatabaseMockMovementsForActor(t *testing.T) {
	repo := &repositoryStub{items: []domain.Movement{{ID: 1, Type: domain.TypeDeposit, Amount: 10000, Prototype: true}}}
	items, err := NewService(repo).List(context.Background(), auth.Actor{UserID: 77})
	if err != nil || repo.userID != 77 || len(items) != 1 || !items[0].Prototype {
		t.Fatalf("List() items=%+v err=%v user=%d", items, err, repo.userID)
	}
}
