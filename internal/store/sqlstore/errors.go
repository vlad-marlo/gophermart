package sqlstore

import "errors"

var (
	ErrInternal           = errors.New("internal server error")
	ErrLoginAlreadyInUse  = errors.New("login already in use")
	ErrUncorrectLoginData = errors.New("uncorrect login data")
)