package store

import (
	"context"

	"github.com/vlad-marlo/gophermart/internal/model"
)

type (
	Storage interface {
		User() UserRepository
		Order() OrderRepository
		Withdraws() WithdrawRepository
		Close()
	}
	UserRepository interface {
		Migrate(ctx context.Context) error
		Create(ctx context.Context, u *model.User) error
		GetByLogin(ctx context.Context, login string) (*model.User, error)
		ExistsWithID(ctx context.Context, id int) bool
		GetBalance(ctx context.Context, id int) (balance *model.UserBalance, err error)
		IncrementBalance(ctx context.Context, id int, add float32) error
	}
	OrderRepository interface {
		Migrate(ctx context.Context) error
		Register(ctx context.Context, user, number int) error
		GetAllByUser(ctx context.Context, user int) (res []*model.Order, err error)
		ChangeStatus(ctx context.Context, user int, m *model.OrderInAccrual) error
		GetUnprocessedOrders(ctx context.Context) ([]*model.OrderInPoll, error)
		ChangeStatusAndIncrementUserBalance(ctx context.Context, user int, m *model.OrderInAccrual) error
	}
	WithdrawRepository interface {
		Migrate(ctx context.Context) error
		Withdraw(ctx context.Context, user int, w *model.Withdraw) error
		GetAllByUser(ctx context.Context, user int) (w []*model.Withdraw, err error)
	}
)
