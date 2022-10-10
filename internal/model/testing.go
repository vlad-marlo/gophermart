package model

import (
	"testing"
)

func TestUser(t *testing.T) *User {
	t.Helper()

	u := &User{
		Login:    "login",
		Password: "password",
	}
	_ = u.BeforeCreate()
	return u
}
