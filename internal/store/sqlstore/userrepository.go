package sqlstore

import (
	"database/sql"

	"github.com/vlad-marlo/gophermart/internal/store/model"
)

type userRepository struct {
	db *sql.DB
}

// newUserRepository ...
func newUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

// Create ...
func (r *userRepository) Create(u *model.User) error {
	return nil
}
