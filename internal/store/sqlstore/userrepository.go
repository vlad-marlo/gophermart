package sqlstore

import (
	"context"
	"database/sql"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
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
func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users(id, login, password) VALUES $1, $2, $3`,
		u.ID,
		u.Login,
		u.EncryptedPassword,
	); err != nil {
		if pgErr := err.(*pq.Error); pgErr.Code == pgerrcode.UniqueViolation {
			return ErrLoginAlreadyInUse
		}
		return err
	}
	return nil
}
