package model

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type (
	User struct {
		ID                int    `json:"-"`
		Login             string `json:"login"`
		Password          string `json:"password"`
		EncryptedPassword string `json:"-"`
	}
	UserBalance struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}
)

// BeforeCreate ...
func (u *User) BeforeCreate() error {
	if len(u.EncryptedPassword) == 0 {
		enc, err := EncryptString(u.Password)
		if err != nil {
			return fmt.Errorf("encrypt string: %v", err)
		}
		u.EncryptedPassword = enc
	}
	return nil
}

func (u *User) Valid() bool {
	if len(u.Login) <= 3 {
		return false
	} else if len(u.Password) <= 5 {
		return false
	}
	return true
}

// EncryptString encryptString ...
func EncryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("gen from pass: %w", err)
	}
	return string(b), nil
}

func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password))
}
