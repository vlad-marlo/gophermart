package sqlstore

import (
	"context"

	"github.com/vlad-marlo/gophermart/internal/model"
)

type withdrawRepository struct {
	s *storage
}

func (r *withdrawRepository) Withdraw(ctx context.Context, w *model.Withdraw) error {
	return nil
}

func (r *withdrawRepository) GetAllByUser(ctx context.Context, user int) ([]*model.Withdraw, error) {
	q := `
	SELECT id
	`
	return nil, nil
}
