package sqlstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type orderRepository struct {
	s *storage
}

func (o *orderRepository) Migrate(ctx context.Context) error {
	q := debugQuery(`
		CREATE TABLE IF NOT EXISTS orders(
			pk BIGSERIAL PRIMARY KEY,
			id BIGINT UNIQUE,
			user_id BIGINT,
			status VARCHAR(50) DEFAULT 'NEW',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			accrual DOUBLE PRECISION DEFAULT 0::DOUBLE PRECISION,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE INDEX IF NOT EXISTS
			index_user_id_orders
		ON orders(user_id);
		CREATE INDEX IF NOT EXISTS
			index_orders_number
		ON orders(id);
	`)

	if _, err := o.s.db.Exec(ctx, q); err != nil {
		return pgError("exec query: %s: %v", err)
	}

	return nil
}

func (o *orderRepository) Register(ctx context.Context, user, number int) error {
	q := debugQuery(`
	INSERT INTO 
		orders(id, user_id)
	VALUES 
		($1, $2);
	`)

	if _, err := o.s.db.Exec(ctx, q, number, user); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return o.getErrByNum(ctx, user, number)
		}
		return pgError("exec: %sum", err)
	}
	return nil
}

func (o *orderRepository) GetAllByUser(ctx context.Context, user int) (orders []*model.Order, err error) {
	q := debugQuery(`
		SELECT 
			x.id, x.status, x.accrual::FLOAT8, x.created_at
		FROM
		    orders x
		WHERE
		    x.user_id = $1
		ORDER BY
		    x.created_at;
	`)

	rows, err := o.s.db.Query(ctx, q, user)
	if err != nil {
		return nil, pgError("query: %sum", err)
	}

	defer rows.Close()

	for rows.Next() {
		var t time.Time
		o := new(model.Order)

		if err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &t); err != nil {
			return nil, pgError("scan rows: %sum", err)
		}
		o.UploadedAt = t.Format(time.RFC3339)
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, pgError("rows err: %sum", err)
	}

	if len(orders) == 0 {
		return nil, store.ErrNoContent
	}

	return orders, nil
}

func (o *orderRepository) getErrByNum(ctx context.Context, user, number int) error {
	q := debugQuery(`
	SELECT EXISTS(
		SELECT
			*
		FROM
			orders
		WHERE
			id = $1 AND user_id = $2
	), EXISTS(
	    SELECT
	        *
	    FROM
	        orders
	    WHERE
	        id = $1
	);`)

	var statusByUser, statusByNum bool
	if err := o.s.db.QueryRow(ctx, q, number, user).Scan(&statusByUser, &statusByNum); err != nil {
		return pgError("query row: %sum", err)
	}

	if statusByUser {
		return store.ErrAlreadyRegisteredByUser
	} else if statusByNum {
		return store.ErrAlreadyRegisteredByAnotherUser
	}

	return nil
}

func (o *orderRepository) ChangeStatus(ctx context.Context, user int, m *model.OrderInAccrual) error {
	q := debugQuery(`
		UPDATE
			orders
		SET
			status = $1,
			accrual = $2::FLOAT8
		WHERE
			id = $3 AND user_id = $4;
	`)

	if _, err := o.s.db.Exec(ctx, q, m.Status, m.Accrual, m.Number, user); err != nil {
		return pgError("exec: %sum", err)
	}
	return nil
}

func (o *orderRepository) GetUnprocessedOrders(ctx context.Context) (res []*model.OrderInPoll, err error) {
	// hardcoded; IDK is it ok
	q := debugQuery(`
		SELECT
		    x.id, x.status, x.user_id
		FROM
		    orders x
		WHERE
		    x.status != 'PROCESSED'
			AND x.status != 'INVALID';
	`)
	q = debugQuery(q)

	rows, err := o.s.db.Query(ctx, q)
	if err != nil {
		return nil, pgError("db query: %sum", err)
	}

	defer rows.Close()

	for rows.Next() {
		o := new(model.OrderInPoll)
		if err := rows.Scan(&o.Number, &o.Status, &o.User); err != nil {
			return nil, pgError("rows scan: %sum", err)
		}
		res = append(res, o)
	}

	if err := rows.Err(); err != nil {
		return nil, pgError("rows err: %sum", err)
	}

	return res, nil
}

func (o *orderRepository) ChangeStatusAndIncrementUserBalance(ctx context.Context, user int, m *model.OrderInAccrual) error {
	qUpdateStatus := debugQuery(`
		UPDATE
			orders
		SET
			status = $1,
			accrual = $2::FLOAT8
		WHERE
			id = $3 AND user_id = $4;
	`)
	qIncrementBalance := `
		UPDATE
			users
		SET
		    balance = balance + $1::DOUBLE PRECISION
		WHERE
		    id = $2;
	`

	tx, err := o.s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %sum", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			o.s.logger.Errorf("update drivers: unable to rollback: %sum", err)
		}
	}()

	if _, err := tx.Exec(ctx, qUpdateStatus, m.Status, m.Accrual, m.Number, user); err != nil {
		return fmt.Errorf("update order: %sum", err)
	}

	if _, err := tx.Exec(ctx, qIncrementBalance, m.Accrual, user); err != nil {
		return fmt.Errorf("increment user balance: %sum", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("update drivers: unable to commmit: %sum", err)
	}

	return nil
}
