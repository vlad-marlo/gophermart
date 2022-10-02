package model

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                int    `json:"-"`
	Login             string `json:"login"`
	Password          string `json:"password"`
	EncryptedPassword string `json:"-"`
}

// NewUser ...

// BeforeCreate ...
func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := EncryptString(u.Password)
		if err != nil {
			return fmt.Errorf("encrypt string: %v", err)
		}
		u.EncryptedPassword, u.Password = enc, ""
	}
	return nil
}

// EncryptString encryptString ...
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
