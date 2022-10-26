package sqlstore_test

import (
	"context"
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
	tt := []struct {
		name    string
		login   string
		wantErr error
	}{
		{
			name:    "good test case #1",
			login:   "login",
			wantErr: nil,
		},
		{
			name:    "duplicate login #1",
			login:   "login",
			wantErr: store.ErrLoginAlreadyInUse,
		},
	}
	testStore, teardown := sqlstore.TestStore(t, conStr)
	defer teardown("users")
	defer logger.DeleteLogFolderAndFile(t)
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			u := model.TestUser(t, tc.login)
			err := testStore.User().Create(context.TODO(), u)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			u1, err := testStore.User().GetByLogin(context.TODO(), u.Login)
			if tc.wantErr == nil {
				require.NoError(t, err)

				if u.Login != u1.Login || u1.ID == 0 {
					t.Fatalf("something went wrong")
				}

				require.True(t, testStore.User().ExistsWithID(context.TODO(), u.ID))
			} else {
				require.False(t, testStore.User().ExistsWithID(context.TODO(), u.ID))
			}
		})

	}
}

// TestUserRepository_GetByLogin_UnExisting ...
func TestUserRepository_GetByLogin_UnExisting(t *testing.T) {
	var tt []string
	for i := 0; i < 10; i++ {
		tt = append(tt, uuid.New().String())
	}

	// init storage for test
	s, teardown := sqlstore.TestStore(t, conStr)
	defer teardown("users")
	defer logger.DeleteLogFolderAndFile(t)
	for _, tc := range tt {
		t.Run(tc, func(t *testing.T) {
			_, err := s.User().GetByLogin(context.TODO(), tc)
			require.ErrorIs(t, err, store.ErrIncorrectLoginData)
		})
	}
}
