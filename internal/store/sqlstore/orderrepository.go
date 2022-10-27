package sqlstore

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type orderRepository struct {
	s *storage
}

func (o *orderRepository) Migrate(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS orders(
			pk BIGSERIAL PRIMARY KEY,
			id BIGINT UNIQUE,
			user_id BIGINT,
			status VARCHAR(50) DEFAULT 'NEW',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			accrual bigint
		);`,

		`ALTER TABLE IF EXISTS
			orders
		ADD CONSTRAINT
			fk_user_order
		FOREIGN KEY (user_id) REFERENCES users(id);`,

		`CREATE INDEX IF NOT EXISTS
			index_user_id_orders
		ON orders(user_id);`,
		`CREATE INDEX IF NOT EXISTS
			index_orders_number
		ON orders(id);`,
	}

	for i, q := range queries {
		o.s.logger.WithFields(logrus.Fields{
			"sql":   debugQuery(q),
			"query": i + 1,
		})
		if _, err := o.s.db.Exec(ctx, q); err != nil { // "42710"
			if pgErr, ok := err.(*pgconn.PgError); !(ok && pgErr.Code == pgerrcode.DuplicateObject) {
				return fmt.Errorf("exec query %d: %v", i+1, pgError(err))
			}
		}
	}

	return nil
}

func (o *orderRepository) Register(ctx context.Context, user, number int) error {
	q := `
		INSERT INTO 
		    orders(id, user_id)
		VALUES 
			($1, $2);
	`
	o.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace(debugQuery(q))

	if _, err := o.s.db.Exec(ctx, q, number, user); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return o.getErrByNum(ctx, user, number)
		}
		return fmt.Errorf("exec: %v", pgError(err))
	}
	return nil
}

func (o *orderRepository) GetAllByUser(ctx context.Context, user int) (res []*model.Order, err error) {
	q := `
		SELECT 
			x.id, x.status, x.accrual::numeric::int, x.created_at
		FROM
		    orders x
		WHERE
		    x.user_id = $1
		ORDER BY
		    x.created_at;
	`
	o.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace(debugQuery(q))

	rows, err := o.s.db.Query(ctx, q, user)
	if err != nil {
		return nil, fmt.Errorf("query: %v", pgError(err))
	}
	for rows.Next() {
		var (
			t time.Time
		)
		o := new(model.Order)
		if err := pgError(rows.Scan(&o.Number, &o.Status, &o.Accrual, &t)); err != nil {
			return nil, fmt.Errorf("scan rows: %v", err)
		}
		o.UploadedAt = t.Format(time.RFC3339)
		res = append(res, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %v", err)
	}
	return res, nil
}

func (o *orderRepository) getErrByNum(ctx context.Context, user, number int) error {
	q := `
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
	        id = $3
	);
	`
	o.s.logger.WithFields(map[string]interface{}{
		"request_id": middleware.GetReqID(ctx),
		"query":      debugQuery(q),
	}).Trace(user, number)

	var statusByUser, statusByNum bool
	if err := o.s.db.QueryRow(ctx, q, number, user, number).Scan(&statusByUser, &statusByNum); err != nil {
		return pgError(err)
	}
	if statusByUser {
		return store.ErrAlreadyRegisteredByUser
	}
	if statusByNum {
		return store.ErrAlreadyRegisteredByAnotherUser
	}
	return nil
}

func (o *orderRepository) ChangeStatus(ctx context.Context, user int, m *model.OrderInAccrual) error {
	q := `
		UPDATE
			orders
		SET
			status = $1,
			accrual = $2
		WHERE
			id = $3 AND user_id = $4;
	`
	o.s.logger.Trace(debugQuery(q))

	if _, err := o.s.db.Exec(ctx, q, m.Status, m.Accrual, m.Number, user); err != nil {
		return fmt.Errorf("exec: %v", pgError(err))
	}
	return nil
}

func (o *orderRepository) GetUnprocessedOrders(ctx context.Context) (res []*model.OrderInPoll, err error) {
	// hardcoded; IDK is it ok
	q := `
		SELECT
		    x.id, x.status, x.user_id
		FROM
		    orders x
		WHERE
		    x.status != 'PROCESSED'
			AND x.status != 'INVALID';
	`
	o.s.logger.Trace(debugQuery(q))
	rows, err := o.s.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("db query: %v", err)
	}

	for rows.Next() {
		o := new(model.OrderInPoll)
		if err := rows.Scan(&o.Number, &o.Status, &o.User); err != nil {
			return nil, fmt.Errorf("rows scan: %v", err)
		}
		res = append(res, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %v", err)
	}

	return res, nil
}
