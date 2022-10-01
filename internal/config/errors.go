package config

import "errors"

var (
	ErrEmptyDataBaseURI = errors.New("DB URI must be not null")
)
