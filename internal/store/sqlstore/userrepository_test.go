package sqlstore_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/uuid"
	"github.com/vlad-marlo/gophermart/pkg/logger"

	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

// TestUserRepository_Create ...
func TestUserRepository_Create(t *testing.T) {
	if conStr == "" {
		t.Skip("conn string is not defined")
	}

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	tt := []struct {
		name    string
		login   string
		wantErr error
	}{
		{
			name:    "good test case #1",
			login:   userLogin1,
			wantErr: nil,
		},
		{
			name:    "good test case #1",
			login:   userLogin2,
			wantErr: nil,
		},
		{
			name:    "duplicate login #1",
			login:   userLogin1,
			wantErr: store.ErrLoginAlreadyInUse,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			u := model.TestUser(t, tc.login)
			err := s.User().Create(context.Background(), u)

			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestUserRepository_GetByLogin(t *testing.T) {
	if conStr == "" {
		t.Skip("conn string is not defined")
	}

	tt := []struct {
		name    string
		login   string
		wantErr error
	}{
		{
			name:    "good test case #1",
			login:   userLogin1,
			wantErr: nil,
		},
		{
			name:    "duplicate login #1",
			login:   userLogin1,
			wantErr: store.ErrLoginAlreadyInUse,
		},
	}

	s, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName)
	defer logger.DeleteLogFolderAndFile(t)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			u := model.TestUser(t, tc.login)
			err := s.User().Create(ctx, u)

			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}

			u1, err := s.User().GetByLogin(ctx, u.Login)
			if tc.wantErr == nil {
				require.NoError(t, err)

				if u.Login != u1.Login || u1.ID == 0 {
					t.Fatalf("something went wrong")
				}

				require.True(t, s.User().ExistsWithID(ctx, u.ID))
			} else {
				require.False(t, s.User().ExistsWithID(ctx, u.ID))
			}
		})

	}
}

// TestUserRepository_GetByLogin_UnExisting ...
func TestUserRepository_GetByLogin_UnExisting(t *testing.T) {
	var tt []string
	if conStr == "" {
		t.Skip("conn string is not defined")
	}

	for i := 0; i < 10; i++ {
		tt = append(tt, uuid.New().String())
	}

	// init storage for test
	s, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName)
	defer logger.DeleteLogFolderAndFile(t)

	for _, tc := range tt {
		t.Run(tc, func(t *testing.T) {
			_, err := s.User().GetByLogin(context.Background(), tc)
			require.ErrorIsf(t, err, store.ErrIncorrectLoginData, "got unexpected error: %v", err)
		})
	}
}

func TestUserRepository_IncrementBalance(t *testing.T) {
	if conStr == "" {
		t.Skip()
	}

	s, teardown := sqlstore.TestStore(t, conStr)
	defer func() {
		teardown(userTableName)
		logger.DeleteLogFolderAndFile(t)
	}()

	ctx := context.Background()
	u := model.TestUser(t, userLogin1)

	err := s.User().Create(context.Background(), u)
	assert.NoErrorf(t, err, "create user %v", err)

	for add := 0.1; add <= 0.2; add += 0.0001 {

		bal, err := s.User().GetBalance(ctx, u.ID)
		assert.NoErrorf(t, err, "get user balance: %v", err)

		err = s.User().IncrementBalance(ctx, u.ID, add)
		assert.NoErrorf(t, err, "increment balance: %v", err)

		after, err := s.User().GetBalance(ctx, u.ID)
		assert.NoErrorf(t, err, "get user balance after incrementation: %v", err)

		assert.True(t, bal.Current+add == after.Current && bal.Withdrawn == after.Withdrawn, "current balance is not correct after incrementation")
	}
}
