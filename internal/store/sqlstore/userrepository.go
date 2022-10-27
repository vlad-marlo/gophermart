package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type userRepository struct {
	s *storage
}

// debugQuery ...
func debugQuery(q string) string {
	q = strings.ReplaceAll(q, "\t", "")
	q = strings.ReplaceAll(q, "\n", " ")
	// this need if anywhere in query used spaces instead of \t
	q = strings.ReplaceAll(q, "    ", "")
	return q
}

// pgError checks err implements postgres error or not. If implements then returns error with postgres format or returns error
func pgError(err error) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return fmt.Errorf(
			"SQL error: %s, Detail: %s, Where: %s, Code: %s, State: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState(),
		)
	}
	return err
}

// Migrate ...
func (r *userRepository) Migrate(ctx context.Context) error {
	q := `
	CREATE TABLE IF NOT EXISTS users(
		id BIGSERIAL UNIQUE PRIMARY KEY NOT NULL,
		login VARCHAR UNIQUE NOT NULL,
		password VARCHAR NOT NULL,
		balance money DEFAULT 0,
		spent INT DEFAULT 0
	);
	`
	if _, err := r.s.db.Exec(ctx, q); err != nil {
		return pgError(err)
	}
	return nil
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

	r.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace(debugQuery(q))

	if err := u.BeforeCreate(); err != nil {
		return fmt.Errorf("before create: %v", err)
	}

	if err := r.s.db.QueryRow(
		ctx,
		q,
		u.Login,
		u.EncryptedPassword,
	).Scan(&u.ID); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return store.ErrLoginAlreadyInUse
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
	r.s.logger.WithFields(logrus.Fields{
		"request_id": id,
	}).Trace(debugQuery(q))

	// we don't need url model, just id
	rows, err := r.s.db.Query(
		ctx,
		q,
		login,
	)

	// closing rows
	defer rows.Close()

	// check error from query context
	if err != nil {
		r.s.logger.WithField("request_id", id).Tracef("err=%s get id by login=%s", err, login)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrIncorrectLoginData
		}
		return nil, fmt.Errorf("query context: %v", pgError(err))
	}
	r.s.logger.WithField("request_id", id).Tracef("get id by login=%s", login)

	// getting data
	if rows.Next() {
		if err := rows.Scan(&u.ID, &u.EncryptedPassword); err != nil {
			return nil, fmt.Errorf("scan: %v", pgError(err))
		}
		return u, nil
	}
	return nil, store.ErrIncorrectLoginData
}

// ExistsWithID ...
func (r *userRepository) ExistsWithID(ctx context.Context, id int) bool {
	var res bool
	q := `
		SELECT EXISTS(
			SELECT
				*
			FROM
				users
			WHERE
				id=$1
		);
	`
	r.s.logger.WithFields(logrus.Fields{
		"request_id": middleware.GetReqID(ctx),
	}).Trace(debugQuery(q))

	if err := r.s.db.QueryRow(
		ctx,
		q,
		id,
	).Scan(&res); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			r.s.logger.WithFields(logrus.Fields{
				"request_id": middleware.GetReqID(ctx),
			}).Errorf("Exists with id: scan: %v", pgError(pgErr))
		}
		return false
	}
	return res
}

// GetBalance ...
func (r *userRepository) GetBalance(ctx context.Context, id int) (balance *model.UserBalance, err error) {
	balance = new(model.UserBalance)
	q := `
		SELECT 
			balance::numeric::float8, spent
		FROM 
			users 
		WHERE 
			id = $1;
	`
	l := r.s.logger.WithFields(logrus.Fields{
		"request_id": middleware.GetReqID(ctx),
	})

	l.Trace(debugQuery(q))
	rows, err := r.s.db.Query(ctx, q, id)
	if err != nil {
		return nil, pgError(err)
	}
	l.Trace("successful get balance")

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %v", err)
	}
	l.Trace("rows err is nil")

	for rows.Next() {
		if err := rows.Scan(&balance.Current, &balance.Withdrawn); err != nil {
			return nil, fmt.Errorf("rows scan: %v", pgError(err))
		}
		l.Trace("return balance")
		return balance, nil
	}

	return nil, store.ErrNoContent
}

// IncrementBalance ...
func (r *userRepository) IncrementBalance(ctx context.Context, id, add int) error {
	q := `
		UPDATE
			users
		SET
			balance = balance + $1::numeric::money
		WHERE
			id = $2;
	`
	if add <= 0 {
		return fmt.Errorf("check args: %v", store.ErrIncorrectData)
	}
	if _, err := r.s.db.Exec(ctx, q, add, id); err != nil {
		return fmt.Errorf("db exec: %v", pgError(err))
	}
	return nil
}
