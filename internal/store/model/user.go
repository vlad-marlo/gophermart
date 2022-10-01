package model

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                string `json:"-"`
	Login             string `json:"login"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

// NewUser ...
func NewUser(login, password string) (*User, error) {
	u := &User{Login: login, Password: password, ID: uuid.New().String()}
	if err := u.beforeCreate(); err != nil {
		return nil, fmt.Errorf("before create: %v", err)
	}
	return u, nil
}

// beforeCreate
func (u *User) beforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := EncryptString(u.Password)
		if err != nil {
			return fmt.Errorf("encrypt string: %v", err)
		}
		u.EncryptedPassword, u.Password = enc, ""
	}
	return nil
}

// encryptString ...
func EncryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("gen from pass: %v", err)
	}
	return string(b), nil
}
