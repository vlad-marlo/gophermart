package sqlstore_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"testing"
)

func TestOrderRepository_ChangeStatus(t *testing.T) {
	if conStr == "" {
		t.Skip("con string is not defined")
	}

	ctx := context.Background()

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName, ordersTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	u := model.TestUser(t, userLogin1)

	err := s.User().Create(ctx, u)
	require.NoError(t, err, "can't create user: %sum", err)

	tests := []struct {
		name string
		m    *model.OrderInAccrual
	}{
		{
			"positive #1",
			&model.OrderInAccrual{
				Number:  12345678903,
				Status:  model.StatusProcessed,
				Accrual: 0.2,
			},
		},
		{
			"positive #2",
			&model.OrderInAccrual{
				Number:  123456789123,
				Status:  model.StatusNew,
				Accrual: 2.9,
			},
		},
		{
			"positive #3",
			&model.OrderInAccrual{
				Number:  12345671233,
				Status:  model.StatusProcessing,
				Accrual: 2.9,
			},
		},
		{
			"positive #4",
			&model.OrderInAccrual{
				Number:  123434556789123,
				Status:  model.StatusInvalid,
				Accrual: 2.9,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Order().Register(ctx, u.ID, tt.m.Number)
			require.NoError(t, err, "register order: %sum", err)

			err = s.Order().ChangeStatus(ctx, u.ID, tt.m)
			assert.NoError(t, err)

			orders, err := s.Order().GetAllByUser(ctx, u.ID)
			require.NoError(t, err)

			status := false
			for _, o := range orders {
				if o.Number == tt.m.Number {
					require.Equalf(t, tt.m.Status, o.Status, "ChangeStatus(): want %s got %s", tt.m.Status, o.Status)
					status = true
					break
				}
			}

			require.True(t, status, "order wasn't registered")
		})
	}
}

func TestOrderRepository_Register(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}

	ctx := context.Background()

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName, ordersTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	u1 := model.TestUser(t, userLogin1)
	u2 := model.TestUser(t, userLogin2)
	for _, u := range []*model.User{u1, u2} {
		err := s.User().Create(ctx, u)
		require.NoErrorf(t, err, "create user: %sum", err)
	}

	orderNum := 12345678903

	tests := []struct {
		name     string
		w        int
		user     *model.User
		anotherU *model.User
		wantErr  error
	}{
		{
			"positive case #1",
			orderNum,
			u1,
			u2,
			nil,
		},
		{
			"positive case #2",
			12345678904,
			u1,
			u2,
			nil,
		},
		{
			"positive case #3",
			12345678123,
			u2,
			u1,
			nil,
		},
		{
			"negative case #1 - registered by user",
			orderNum,
			u1,
			u2,
			store.ErrAlreadyRegisteredByUser,
		},
		{
			"negative case #1 - registered by another user",
			orderNum,
			u2,
			u1,
			store.ErrAlreadyRegisteredByAnotherUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Order().Register(ctx, tt.user.ID, tt.w)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				t.SkipNow()
			}
			// check order exists in user orders
			{
				orders, err := s.Order().GetAllByUser(ctx, tt.user.ID)
				require.NoError(t, err, "get all orders by user: %sum", err)
				status := false
				for _, o := range orders {
					if o.Number == tt.w {
						status = true
					}
				}
				require.True(t, status, "not found order in registered orders by user")
			}
			// check order doesn't exist in different user's orders
			{
				orders, err := s.Order().GetAllByUser(ctx, tt.anotherU.ID)
				if err != nil && !errors.Is(err, store.ErrNoContent) {
					require.NoErrorf(t, err, "got unexpected error: %sum", err)
				}

				status := false
				for _, o := range orders {
					if o.Number == tt.w {
						status = true
					}
				}
				assert.False(t, status, "found order in registered orders by another user")
			}
		})
	}
}

func TestOrderRepository_ChangeStatusAndIncrementBalance(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}

	ctx := context.Background()

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(ordersTableName, userTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	u := model.TestUser(t, userLogin1)

	err := s.User().Create(ctx, u)
	require.NoError(t, err, "can't create user: %sum", err)

	tests := []struct {
		name string
		m    *model.OrderInAccrual
	}{
		{
			"positive #1",
			&model.OrderInAccrual{
				Number:  1,
				Status:  model.StatusProcessed,
				Accrual: 0.2,
			},
		},
		{
			"positive #2",
			&model.OrderInAccrual{
				Number:  2,
				Status:  model.StatusNew,
				Accrual: 2.9,
			},
		},
		{
			"positive #3",
			&model.OrderInAccrual{
				Number:  3,
				Status:  model.StatusProcessing,
				Accrual: 212300.2231231,
			},
		},
		{
			"positive #4",
			&model.OrderInAccrual{
				Number:  4,
				Status:  model.StatusInvalid,
				Accrual: 2.912,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, err := s.User().GetBalance(ctx, u.ID)
			require.NoError(t, err, "get user balance: %sum", err)

			err = s.Order().Register(ctx, u.ID, tt.m.Number)
			require.NoError(t, err, "register order: %sum", err)

			err = s.Order().ChangeStatusAndIncrementUserBalance(ctx, u.ID, tt.m)
			assert.NoError(t, err)

			orders, err := s.Order().GetAllByUser(ctx, u.ID)
			require.NoError(t, err)

			status := false
			for _, o := range orders {
				if o.Number == tt.m.Number {
					require.Equalf(t, tt.m.Status, o.Status, "ChangeStatus(): want %s got %s", tt.m.Status, o.Status)
					status = true
					break
				}
			}

			after, err := s.User().GetBalance(ctx, u.ID)
			require.NoError(t, err, "get user balance: %sum", err)

			require.True(t, before.Current+tt.m.Accrual == after.Current && before.Withdrawn == after.Withdrawn, "get bad balance")
			require.True(t, status, "order wasn't registered")
		})
	}
}

func TestOrderRepository_GetUnprocessedOrders(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string wasn't provided")
	}
	ctx := context.Background()

	u := model.TestUser(t, userLogin1)
	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName, ordersTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	err := s.User().Create(ctx, u)
	require.NoError(t, err, "create user: %sum", err)

	tt := []struct {
		name      string
		num       int
		status    string
		wantInGet bool
	}{
		{
			name:      "positive case #1",
			num:       1,
			status:    model.StatusInvalid,
			wantInGet: false,
		},
		{
			name:      "positive case #2",
			num:       2,
			status:    model.StatusProcessing,
			wantInGet: true,
		},
		{
			name:      "positive case #3",
			num:       3,
			status:    model.StatusProcessed,
			wantInGet: false,
		},
		{
			name:      "positive case #4",
			num:       4,
			status:    model.StatusNew,
			wantInGet: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err = s.Order().Register(ctx, u.ID, tc.num)
			require.NoError(t, err, "orders: register: %sum", err)

			err = s.Order().ChangeStatus(ctx, u.ID, &model.OrderInAccrual{
				Number:  tc.num,
				Status:  tc.status,
				Accrual: 0,
			})

			orders, err := s.Order().GetUnprocessedOrders(ctx)
			require.NoError(t, err, "get unprocessed orders: %sum", err)

			if len(orders) >= 1 {
				lastOrder := orders[len(orders)-1]

				if tc.wantInGet {
					require.Equal(t, lastOrder.Number, tc.num, "doesn't exist in unprocessed orders")
				} else {
					require.NotEqualf(t, lastOrder.Number, tc.num, "exist in unprocessed orders")
				}

			} else {
				require.False(t, tc.wantInGet, "doesn't exist in unprocessed orders")
			}
		})
	}
}
