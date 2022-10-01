package store

import "github.com/vlad-marlo/gophermart/internal/store/model"

type (
	Storage interface {
		User() UserRepository
		Close() error
	}
	UserRepository interface {
		Create(u *model.User) error
	}
)
