package store

import "errors"

var (
	ErrLoginAlreadyInUse              = errors.New("login already in use")
	ErrIncorrectLoginData             = errors.New("uncorrect login data")
	ErrIncorrectData                  = errors.New("uncorrect arguments")
	ErrAlreadyRegisteredByUser        = errors.New("registered by user")
	ErrAlreadyRegisteredByAnotherUser = errors.New("registered by another user")
	ErrNoContent                      = errors.New("no data to return")
	ErrPaymentRequired                = errors.New("payment required")
)
