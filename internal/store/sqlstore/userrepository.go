package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type userRepository struct {
	db *sql.DB
	l  *logrus.Logger
}

// Create ...
func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users(id, login, password) VALUES ($1, $2, $3);`,
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

// GetByLogin GetIDByLoginAndPass ...
func (r *userRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	u := &model.User{Login: login}

	// we don't need url model, just id
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, password FROM users WHERE login=$1;`,
		login,
	)
	defer func() {
		if err := rows.Close(); err != nil {
			r.l.Warnf("user repo: get by login: rows close: %v", err)
		}
	}()
	if err != nil {
		r.l.Tracef("err=%s get id by login=%s", err, login)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUncorrectLoginData
		}
		return nil, fmt.Errorf("query context: %v", err)
	}

	if rows.Next() {
		if err := rows.Scan(&u.ID, &u.EncryptedPassword); err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}
		return u, nil
	}
	return nil, ErrUncorrectLoginData
}

// ExistsWithID ...
func (r *userRepository) ExistsWithID(ctx context.Context, id int) bool {
	var res bool
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT * FROM users WHERE id=$1);`,
		id,
	).Scan(&res); err != nil {
		r.l.Warnf("user repo: exists with id: scan: %v", err)
		return false
	}
	return res
}
