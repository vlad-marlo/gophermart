package store

import (
	"context"

	"github.com/vlad-marlo/gophermart/internal/model"
)

type (
	Storage interface {
		User() UserRepository
		Close() error
		Order() OrderRepository
	}
	UserRepository interface {
		Create(ctx context.Context, u *model.User) error
		GetByLogin(ctx context.Context, login string) (*model.User, error)
		ExistsWithID(ctx context.Context, id int) bool
		GetBalance(ctx context.Context, id int) (balance *model.UserBalance, err error)
	}
	OrderRepository interface {
	}
)
