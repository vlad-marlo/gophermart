package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/store/model"
)

type userRepository struct {
	db *sql.DB
	l  *logrus.Logger
}

// Create ...
func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users(user_id, login, password) VALUES ($1, $2, $3);`,
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
		`SELECT user_id FROM users WHERE login=$1 AND password=$2`,
		login,
		enc,
	)
	defer rows.Close()
	if err != nil {
		r.l.Debugf("err=%s get id by login=%s", err, login)
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUncorrectLoginData
		}
		return "", fmt.Errorf("query context: %v", err)
	}

	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", fmt.Errorf("scan: %v", err)
		}
		return id, nil
	}
	return "", ErrUncorrectLoginData
}

// ExistsWithID ...
func (r *userRepository) ExistsWithID(ctx context.Context, id string) (bool, error) {
	var res bool
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT * FROM urls WHERE user_id=$1);`,
		id,
	).Scan(&res); err != nil {
		return false, fmt.Errorf("query: %v", err)
	}
	return res, nil
}
