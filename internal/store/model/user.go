package model

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                int    `json:"-"`
	Login             string `json:"login"`
	Password          string `json:"password"`
	EncryptedPassword string `json:"-"`
}

// NewUser ...
func NewUser(login, password string) (*User, error) {
	u := &User{Login: login, Password: password}
	if err := u.BeforeCreate(); err != nil {
		return nil, fmt.Errorf("before create: %v", err)
	}
	return u, nil
}

// BeforeCreate
func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := EncryptString(u.Password)
		if err != nil {
			return fmt.Errorf("encrypt string: %v", err)
		}
		u.EncryptedPassword, u.Password = enc, ""
	}
	if len(u.ID) == 0 {
		u.ID = uuid.New()
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

func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password))
}
