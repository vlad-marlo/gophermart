package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/internal/store/model"
)

type userRepository struct {
	db *sql.DB
}

func newUserRepository(db *sql.DB) store.UserRepository {
	return &userRepository{db}
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

// GetIDByLoginAndPass ...
func (r *userRepository) GetIDByLoginAndPass(ctx context.Context, login, pass string) (string, error) {
	var id string

	enc, err := model.EncryptString(pass)
	if err != nil {
		return "", fmt.Errorf("encrypt pass: %v", err)
	}

	// we don't need url model, just id
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT users(id) WHERE login=$1 AND password=$2`,
		login,
		enc,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUncorrectLoginData
		}
		return "", fmt.Errorf("query context: %v", err)
	}

	// I didn't use rows.Next() because we will get 0 or 1 rows constantly. That means if we didn't get
	// noRows error that we have one row which we will scan
	if err := rows.Scan(&id); err != nil {
		return "", fmt.Errorf("scan: %v", err)
	}

	return id, nil
}
