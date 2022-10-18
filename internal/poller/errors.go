package poller

import "errors"

var (
	ErrInternal        = errors.New("internal server error")
	ErrTooManyRequests = errors.New("internal server error")
)
