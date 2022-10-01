package store

import (
	"context"

	"github.com/vlad-marlo/gophermart/internal/store/model"
)

type (
	Storage interface {
		User() UserRepository
		Close() error
	}
	UserRepository interface {
		Create(ctx context.Context, u *model.User) error
	}
)
