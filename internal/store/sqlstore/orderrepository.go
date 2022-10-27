package sqlstore

import (
	"context"
	"time"

	"github.com/lib/pq"

	"github.com/jackc/pgerrcode"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type orderRepository struct {
	s *storage
}

func (o *orderRepository) Migrate(ctx context.Context) error {
	q := `
		CREATE TABLE IF NOT EXISTS orders(
			pk BIGSERIAL PRIMARY KEY,
			id BIGINT UNIQUE,
			user_id BIGINT,
			status VARCHAR(50) DEFAULT 'NEW',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			accrual float4,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE INDEX IF NOT EXISTS
			index_user_id_orders
		ON orders(user_id);
		CREATE INDEX IF NOT EXISTS
			index_orders_number
		ON orders(id);
	`

	if _, err := o.s.db.Exec(ctx, q); err != nil {
		return pgError("exec query: %s: %v", err)
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

	if _, err := o.s.db.Exec(ctx, q, number, user); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return o.getErrByNum(ctx, user, number)
		}
		return pgError("exec: %v", err)
	}
	return nil
}

func (o *orderRepository) GetAllByUser(ctx context.Context, user int) (orders []*model.Order, err error) {
	// TODO: FIX 500 ERR
	q := `
		SELECT 
			x.id, x.status, x.accrual, x.created_at
		FROM
		    orders x
		WHERE
		    x.user_id = $1 AND x.accrual IS NOT NULL
		ORDER BY
		    x.created_at;
	`

	rows, err := o.s.db.Query(ctx, q, user)
	if err != nil {
		return nil, pgError("query: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var t time.Time
		o := new(model.Order)

		if err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &t); err != nil {
			return nil, pgError("scan rows: %v", err)
		}
		o.UploadedAt = t.Format(time.RFC3339)
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, pgError("rows err: %v", err)
	}

	return orders, nil
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
	        id = $1
	);`

	var statusByUser, statusByNum bool
	if err := o.s.db.QueryRow(ctx, q, number, user).Scan(&statusByUser, &statusByNum); err != nil {
		return pgError("query row: %v", err)
	}

	if statusByUser {
		return store.ErrAlreadyRegisteredByUser
	} else if statusByNum {
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

	if _, err := o.s.db.Exec(ctx, q, m.Status, m.Accrual, m.Number, user); err != nil {
		return pgError("exec: %v", err)
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

	rows, err := o.s.db.Query(ctx, q)
	if err != nil {
		return nil, pgError("db query: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		o := new(model.OrderInPoll)
		if err := rows.Scan(&o.Number, &o.Status, &o.User); err != nil {
			return nil, pgError("rows scan: %v", err)
		}
		res = append(res, o)
	}

	if err := rows.Err(); err != nil {
		return nil, pgError("rows err: %v", err)
	}

	return res, nil
}
