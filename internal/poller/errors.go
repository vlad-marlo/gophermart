package poller

import "errors"

var (
	ErrInternal         = errors.New("internal server error")
	ErrTooManyRequests  = errors.New("too many requests")
	ErrUnexpectedStatus = errors.New("got unexpected status")
)
