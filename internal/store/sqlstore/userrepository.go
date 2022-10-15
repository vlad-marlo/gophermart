package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/vlad-marlo/gophermart/pkg/logger"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"
)

type userRepository struct {
	db *sql.DB
	l  logger.Logger
}

// debugQuery ...
func debugQuery(q string) string {
	q = strings.ReplaceAll(q, "\t", "")
	q = strings.ReplaceAll(q, "\n", " ")
	// this need if anywhere in query used spaces instead of \t
	q = strings.ReplaceAll(q, "    ", " ")
	return q
}

// pgError checks err implements pq error or not. If implements then returns error with postgres format or returns error
func pgError(err error) error {
	if pgErr, ok := err.(*pq.Error); ok {
		return fmt.Errorf(
			"SQL error: %s, Detail: %s, Where: %s, Code: %s, State: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState(),
		)
	}
	return err
}

// Create ...
func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	q := `
		INSERT INTO
			users(login, password)
		VALUES
			($1, $2)
		RETURNING id;
	`

	r.l.WithFields(logrus.Fields{
		"request_id": middleware.GetReqID(ctx),
		"args": struct {
			User *model.User `json:"user"`
		}{u},
	}).Trace(debugQuery(q))

	if err := u.BeforeCreate(); err != nil {
		return fmt.Errorf("before create: %v", err)
	}

	if err := r.db.QueryRowContext(
		ctx,
		q,
		u.Login,
		u.EncryptedPassword,
	).Scan(&u.ID); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrLoginAlreadyInUse
		}
		return pgError(err)
	}

	return nil
}

// GetByLogin GetIDByLoginAndPass ...
func (r *userRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	q := `
		SELECT
			x.id, x.password
		FROM users AS x
		WHERE x.login=$1;
	`
	u := &model.User{Login: login}
	id := middleware.GetReqID(ctx)

	// trace request
	r.l.WithFields(logrus.Fields{
		"request_id": id,
		"args": struct {
			Login string `json:"login"`
		}{login},
	}).Trace(debugQuery(q))

	// we don't need url model, just id
	rows, err := r.db.QueryContext(
		ctx,
		q,
		login,
	)

	// closing rows
	defer func() {
		if err := rows.Close(); err != nil {
			r.l.WithFields(logrus.Fields{
				"request_id": id,
			}).Errorf("user repo: get by login: rows close: %v", pgError(err))
		}
	}()

	// check error from query context
	if err != nil {
		r.l.WithField("request_id", id).Tracef("err=%s get id by login=%s", err, login)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUncorrectLoginData
		}
		return nil, fmt.Errorf("query context: %v", pgError(err))
	}
	r.l.WithField("request_id", id).Tracef("get id by login=%s", login)

	// getting data
	if rows.Next() {
		if err := rows.Scan(&u.ID, &u.EncryptedPassword); err != nil {
			return nil, fmt.Errorf("scan: %v", pgError(err))
		}
		return u, nil
	}
	return nil, ErrUncorrectLoginData
}

// ExistsWithID ...
func (r *userRepository) ExistsWithID(ctx context.Context, id int) bool {
	var res bool
	q := `
		SELECT EXISTS(
			SELECT
				x.*
			FROM
				users AS x
			WHERE
				x.id=$1
		);
	`
	r.l.WithFields(logrus.Fields{
		"request_id": middleware.GetReqID(ctx),
		"args": struct {
			ID int `json:"id"`
		}{id},
	}).Trace(debugQuery(q))

	if err := r.db.QueryRowContext(
		ctx,
		q,
		id,
	).Scan(&res); err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			r.l.WithFields(logrus.Fields{
				"request_id": middleware.GetReqID(ctx),
			}).Errorf("Exists with id: scan: %v", pgError(pgErr))
		}
		return false
	}
	return res
}

// GetBalance ...
func (r *userRepository) GetBalance(ctx context.Context, id int) (balance *float32, spent *int, err error) {
	q := `
		SELECT 
			x.balance, x.spent 
		FROM 
			users x 
		WHERE 
			x.id = $1;
	`
	r.l.WithFields(logrus.Fields{
		"request_id": middleware.GetReqID(ctx),
		"args": struct {
			ID int `json:"id"`
		}{id},
	}).Trace(debugQuery(q))

	rows, err := r.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, nil, pgError(err)
	}

	defer func() {
		if err := pgError(rows.Close()); err != nil {
			r.l.Errorf("get balance: defer func: %v", err)
		}
	}()

	for rows.Next() {
		if err := pgError(rows.Scan(&balance, &spent)); err != nil {
			return nil, nil, err
		}
		return balance, spent, nil
	}
	return nil, nil, sql.ErrNoRows
}
