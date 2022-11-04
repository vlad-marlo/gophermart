package poller

import "errors"

var (
	ErrInternal        = errors.New("internal server error")
	ErrTooManyRequests = errors.New("too many requests")
	ErrNotFound        = errors.New("not found")
	ErrNoContent       = errors.New("no content")
)
