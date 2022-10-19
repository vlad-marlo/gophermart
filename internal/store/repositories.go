package store

import (
	"context"

	"github.com/vlad-marlo/gophermart/internal/model"
)

type (
	Storage interface {
		User() UserRepository
		Close()
		Order() OrderRepository
	}
	UserRepository interface {
		Create(ctx context.Context, u *model.User) error
		GetByLogin(ctx context.Context, login string) (*model.User, error)
		ExistsWithID(ctx context.Context, id int) bool
		GetBalance(ctx context.Context, id int) (balance *model.UserBalance, err error)
		IncrementBalance(ctx context.Context, id, add int) error
		UseBalance(ctx context.Context, id, use int) error
	}
	OrderRepository interface {
		Register(ctx context.Context, user int, number int) error
		GetAllByUser(ctx context.Context, user int) ([]*model.Order, error)
		ChangeStatus(ctx context.Context, m *model.OrderInAccrual) error
	}
	WithdrawRepository interface {
	}
)
