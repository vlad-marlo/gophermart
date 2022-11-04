package store

import (
	"context"

	"github.com/vlad-marlo/gophermart/internal/model"
)

type (
	// Storage ...
	Storage interface {
		// User ...
		User() UserRepository
		// Order ...
		Order() OrderRepository
		// Withdraws ...
		Withdraws() WithdrawRepository
		// Close ...
		Close()
	}
	// UserRepository ...
	UserRepository interface {
		// Migrate database to current scheme
		Migrate(ctx context.Context) error
		// Create record about user u to storage; could return error if user already exists or other internal error
		Create(ctx context.Context, u *model.User) error
		// GetByLogin search record about user with login and return it if record exits
		GetByLogin(ctx context.Context, login string) (*model.User, error)
		// ExistsWithID check existing record about user with current id or not
		ExistsWithID(ctx context.Context, id int) bool
		// GetBalance return user balance and sum of all user withdrawals
		GetBalance(ctx context.Context, id int) (balance *model.UserBalance, err error)
		// IncrementBalance is adding balance to user with id
		IncrementBalance(ctx context.Context, id int, add float64) error
	}
	OrderRepository interface {
		// Migrate database to current scheme
		Migrate(ctx context.Context) error
		// Register create record about order with unique id which is number
		Register(ctx context.Context, user, number int) error
		// GetAllByUser returns all orders which was registered by user
		GetAllByUser(ctx context.Context, user int) (res []*model.Order, err error)
		// ChangeStatus is changing status of order with id m.Number to status m.Status
		ChangeStatus(ctx context.Context, user int, m *model.OrderInAccrual) error
		// GetUnprocessedOrders return all orders which status is not final('NEW', 'PROCESSING')
		GetUnprocessedOrders(ctx context.Context) ([]*model.OrderInPoll, error)
		// ChangeStatusAndIncrementUserBalance is changing status of order with id m.Number to status m.Status and
		// adding m.Accrual to user balance
		ChangeStatusAndIncrementUserBalance(ctx context.Context, user int, m *model.OrderInAccrual) error
	}
	WithdrawRepository interface {
		// Migrate database to current scheme
		Migrate(ctx context.Context) error
		// Withdraw create record about withdraw and writes-down user balance
		Withdraw(ctx context.Context, user int, w *model.Withdraw) error
		// GetAllByUser return all withdraw records which was created by user
		GetAllByUser(ctx context.Context, user int) (w []*model.Withdraw, err error)
	}
)
