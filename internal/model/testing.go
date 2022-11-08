package model

import (
	"testing"
)

func TestUser(t *testing.T, login string) *User {
	t.Helper()

	u := &User{
		Login:    login,
		Password: "password",
	}
	_ = u.BeforeCreate()
	return u
}

func TestWithdraw(t *testing.T, order int, sum float64) *Withdraw {
	t.Helper()
	return &Withdraw{
		Order: order,
		Sum:   sum,
	}
}

func TestOrder(t *testing.T, number int, status string) *Order {
	t.Helper()
	return &Order{
		Number: number,
		Status: status,
	}
}
