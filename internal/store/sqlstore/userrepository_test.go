package sqlstore_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

func TestUserRepository_Create(t *testing.T) {
	db, teardown := sqlstore.TestStore(t, conStr)
	defer teardown("users")
	u := model.TestUser(t)
	require.NoError(t, db.User().Create(context.TODO(), u))
	u1, err := db.User().GetByLogin(context.TODO(), u.Login)
	require.NoError(t, err)
	if u.Login != u1.Login || u1.ID == 0 {
		t.Fatalf("something went wrong")
	}
}
