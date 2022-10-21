package sqlstore

import "errors"

var (
	ErrLoginAlreadyInUse              = errors.New("login already in use")
	ErrUncorrectLoginData             = errors.New("uncorrect login data")
	ErrUncorrectData                  = errors.New("uncorrect arguments")
	ErrAlreadyRegisteredByUser        = errors.New("registered by user")
	ErrAlreadyRegisteredByAnotherUser = errors.New("registered by another user")
	ErrNoContent                      = errors.New("no data to return")
)
