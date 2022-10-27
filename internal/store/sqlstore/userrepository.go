package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type userRepository struct {
	s *storage
}

// Migrate ...
func (r *userRepository) Migrate(ctx context.Context) error {
	q := `
	CREATE TABLE IF NOT EXISTS users(
		id BIGSERIAL UNIQUE PRIMARY KEY NOT NULL,
		login VARCHAR UNIQUE NOT NULL,
		password VARCHAR NOT NULL,
		balance money DEFAULT 0
	);
	`
	if _, err := r.s.db.Exec(ctx, q); err != nil {
		return sqlErr("exec: %v", err, q)
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

	if err := u.BeforeCreate(); err != nil {
		return fmt.Errorf("before create: %v", err)
	}

	if err := r.s.db.QueryRow(
		ctx,
		q,
		u.Login,
		u.EncryptedPassword,
	).Scan(&u.ID); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return store.ErrLoginAlreadyInUse
		}
		return sqlErr("scan: %v", err, q)
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

	rows, err := r.s.db.Query(
		ctx,
		q,
		login,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrIncorrectLoginData
		}
		return nil, sqlErr("query context: %v", err, q)
	}

	// check error from query context
	// closing rows
	defer rows.Close()

	// getting data
	if rows.Next() {
		if err := rows.Scan(&u.ID, &u.EncryptedPassword); err != nil {
			return nil, sqlErr("rows scan: %v", err, q)
		}
		return u, nil
	}
	if err := rows.Err(); err != nil {
		return nil, sqlErr("rows err: %v", err, q)
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

	if err := r.s.db.QueryRow(
		ctx,
		q,
		id,
	).Scan(&res); err != nil {
		r.s.logger.WithFields(logrus.Fields{
			"request_id": middleware.GetReqID(ctx),
			"sql":        debugQuery(q),
		}).Errorf("exists with id: scan: %v", pgError(err))
		return false
	}
	return res
}

// GetBalance ...
func (r *userRepository) GetBalance(ctx context.Context, id int) (balance *model.UserBalance, err error) {
	q := `
		SELECT
			balance::numeric::float4, (
				SELECT
				    SUM(order_sum)
				FROM
				    withdrawals
				WHERE
				    user_id = $1
			)
		FROM
			users
		WHERE
			id = $1;
	`
	balance = new(model.UserBalance)

	rows, err := r.s.db.Query(ctx, q, id)
	if err != nil {
		return nil, sqlErr("exec query: %v", err, q)
	}

	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&balance.Current, &balance.Withdrawn); err != nil {
			return nil, sqlErr("rows scan: %v", err, q)
		}
		return balance, nil
	}

	if err := rows.Err(); err != nil {
		return nil, sqlErr("rows err: %v", err, q)
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
		return sqlErr("db exec: %v", err, q)
	}
	return nil
}
