package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
)

type userRepository struct {
	db *sql.DB
	l  logger.Logger
}

// Create ...
func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users(login, password) VALUES ($1, $2);`,
		u.Login,
		u.EncryptedPassword,
	); err != nil {
		if pgErr := err.(*pq.Error); pgErr.Code == pgerrcode.UniqueViolation {
			return ErrLoginAlreadyInUse
		}
		return err
	}
	r.l.WithField("request_id", middleware.GetReqID(ctx)).Trace("successful created user with login ", u.Login)
	return nil
}

// GetByLogin GetIDByLoginAndPass ...
func (r *userRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	u := &model.User{Login: login}
	id := middleware.GetReqID(ctx)

	// we don't need url model, just id
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, password FROM users WHERE login=$1;`,
		login,
	)
	defer func() {
		if err := rows.Close(); err != nil {
			r.l.WithFields(logrus.Fields{
				"request_id": id,
			}).Warnf("user repo: get by login: rows close: %v", err)
		}
	}()
	if err != nil {
		r.l.WithField("request_id", id).Tracef("err=%s get id by login=%s", err, login)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUncorrectLoginData
		}
		return nil, fmt.Errorf("query context: %v", err)
	}
	r.l.WithField("request_id", id).Tracef("get id by login=%s", login)

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

	query := `SELECT EXISTS(SELECT * FROM users WHERE id=$1);`
	if err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(&res); err != nil {
		r.l.WithFields(logrus.Fields{
			"sql":        fmt.Sprint(query, id),
			"request_id": middleware.GetReqID(ctx),
		}).Debugf("user repo: exists with id: scan: %v", err)
		return false
	}
	r.l.WithFields(logrus.Fields{
		"sql":        query,
		"request_id": middleware.GetReqID(ctx),
	}).Tracef("res: %v", res)
	return res
}
