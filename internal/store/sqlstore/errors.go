package sqlstore

import "errors"

var (
	ErrLoginAlreadyInUse  = errors.New("login already in use")
	ErrUncorrectLoginData = errors.New("uncorrect login data")
	ErrUncorrectData      = errors.New("uncorrect arguments")
)
