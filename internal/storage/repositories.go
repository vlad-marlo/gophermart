package storage

import "github.com/vlad-marlo/gophermart/internal/storage/model"

type (
	Storage interface {
		User() UserRepository
		Close() error
	}
	UserRepository interface {
		Create(u *model.User) error
	}
)
