package sqlstore_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"github.com/vlad-marlo/gophermart/pkg/luhn"
	"testing"
)

func TestWithdrawalRepository_WithdrawPositive(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string was not provided")
	}

	ctx := context.Background()
	u := model.TestUser(t, userLogin1)

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName, ordersTableName, withdrawalsTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	err := s.User().Create(ctx, u)
	require.NoErrorf(t, err, "create user: %v", err)

	o := model.TestOrder(t, orderNum1, model.StatusNew)
	err = s.Order().Register(ctx, u.ID, o.Number)
	require.NoErrorf(t, err, "register balance: %v", err)

	for withdraw := 4990.0; withdraw < 4999.99; withdraw += 0.01 {

		err = s.User().IncrementBalance(ctx, u.ID, 10000.0)
		require.NoErrorf(t, err, "increment balance: %v", err)

		before, err := s.User().GetBalance(ctx, u.ID)
		require.NoErrorf(t, err, "get user balance: %v", err)

		w := model.TestWithdraw(t, o.Number, withdraw)

		err = s.Withdraws().Withdraw(ctx, u.ID, w)
		require.NoErrorf(t, err, "withdraw user: %v", err)

		after, err := s.User().GetBalance(ctx, u.ID)
		require.NoErrorf(t, err, "get user balance: %v", err)
		msg := fmt.Sprintf(
			"%f + %f = %f || %f + %f = %f",
			before.Withdrawn,
			before.Current,
			before.Current+before.Withdrawn,
			after.Current, after.Withdrawn,
			after.Withdrawn+after.Current,
		)
		require.True(t, before.Withdrawn+before.Current <= after.Withdrawn+after.Current+0.01, msg)
		require.Equal(t, before.Withdrawn+withdraw, after.Withdrawn, "withdraw sum not equal")
		require.Equal(t, before.Current-withdraw, after.Current, "current bal not equal")
	}
}

func TestWithdrawalRepository_WithdrawNegative(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string was not provided")
	}

	ctx := context.Background()
	u := model.TestUser(t, userLogin1)

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName, ordersTableName, withdrawalsTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	err := s.User().Create(ctx, u)
	require.NoErrorf(t, err, "create user: %v", err)

	o := model.TestOrder(t, orderNum1, model.StatusNew)
	err = s.Order().Register(ctx, u.ID, o.Number)
	require.NoErrorf(t, err, "register balance: %v", err)
	withdrawSum := 0.0

	for withdraw := 9999.95; withdraw < 10000.0; withdraw += 0.001 {
		err = s.User().IncrementBalance(ctx, u.ID, 5000.0)
		require.NoErrorf(t, err, "increment balance: %v", err)

		before, err := s.User().GetBalance(ctx, u.ID)
		require.NoErrorf(t, err, "get user balance: %v", err)

		w := model.TestWithdraw(t, o.Number, withdraw)

		err = s.Withdraws().Withdraw(ctx, u.ID, w)
		require.ErrorIs(t, err, store.ErrPaymentRequired, err)

		w.Sum = before.Current
		withdrawSum += before.Current

		err = s.Withdraws().Withdraw(ctx, u.ID, w)
		require.NoError(t, err)

		after, err := s.User().GetBalance(ctx, u.ID)
		require.NoErrorf(t, err, "get user balance: %v", err)

		assert.Equal(t, 0.0, after.Current)
		assert.Equal(t, withdrawSum, after.Withdrawn)
	}
}

func TestWithdrawalsRepository_GetAllByUser(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}

	ctx := context.Background()

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName, ordersTableName, withdrawalsTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	u1 := model.TestUser(t, userLogin1)
	u2 := model.TestUser(t, userLogin2)
	for _, u := range []*model.User{u1, u2} {
		err := s.User().Create(ctx, u)
		require.NoErrorf(t, err, "create user: %s", err)
	}

	orderNum := 12345678903
	err := s.Order().Register(ctx, u1.ID, orderNum)
	require.NoErrorf(t, err, "register order: %s", err)

	tests := []struct {
		name     string
		sum      float64
		order    int
		user     *model.User
		anotherU *model.User
		wantErr  error
	}{
		{
			"positive case #1",
			123.212123,
			orderNum1,
			u1,
			u2,
			nil,
		},
		{
			"positive case #2",
			1231231.232,
			orderNum2,
			u1,
			u2,
			nil,
		},
		{
			"positive case #3",
			2.002,
			orderNum3,
			u2,
			u1,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = s.User().IncrementBalance(ctx, tt.user.ID, tt.sum)
			bal, err := s.User().GetBalance(ctx, tt.user.ID)
			t.Logf("withdraw=%f, got=%f, u=%d", tt.sum, bal.Current, tt.user.ID)
			require.NoError(t, err)
			require.NoErrorf(t, err, "increment user balance: %w", err)
			err = s.Withdraws().Withdraw(ctx, tt.user.ID, model.TestWithdraw(t, tt.order, tt.sum))
			if luhn.Valid(tt.order) {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, store.ErrIncorrectData, err)
				return
			}
			// check order exists in user orders
			{
				withdraws, err := s.Withdraws().GetAllByUser(ctx, tt.user.ID)
				require.NoError(t, err, "get all withdraws by user: %s", err)
				status := false
				for _, o := range withdraws {
					if o.Sum == tt.sum {
						status = true
					}
				}
				require.True(t, status, "not found withdraw in registered orders by user")
			}
			// check order doesn't exist in different user's orders
			{
				orders, err := s.Withdraws().GetAllByUser(ctx, tt.anotherU.ID)
				if err != nil && !errors.Is(err, store.ErrNoContent) {
					require.NoErrorf(t, err, "got unexpected error: %s", err)
				}

				status := false
				for _, o := range orders {
					if o.Sum == tt.sum {
						status = true
					}
				}
				assert.False(t, status, "found order in registered orders by another user")
			}
		})
	}
}
