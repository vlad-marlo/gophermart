package sqlstore

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vlad-marlo/gophermart/internal/model"
)

type orderRepository struct {
	s *storage
}

func (o *orderRepository) Register(ctx context.Context, user, number int) error {
	q := `
		INSERT INTO 
		    orders(id, user_id)
		VALUES 
			($1, $2);
	`
	o.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace(debugQuery(q))

	if _, err := o.s.db.Exec(ctx, q, user, number); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return o.getErrByNum(ctx, user, number)
			}
		}
		return fmt.Errorf("exec: %v", pgError(err))
	}
	return nil
}

func (o *orderRepository) GetAllByUser(ctx context.Context, user int) (res []*model.Order, err error) {
	q := `
		SELECT 
			x.id, x.status, x.accrual, x.created_at
		FROM
		    orders x
		WHERE
		    x.user_id = $1
		ORDER BY
		    x.created_at ASC;
	`
	o.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace(debugQuery(q))

	rows, err := o.s.db.Query(ctx, q, user)
	if err != nil {
		return nil, fmt.Errorf("query: %v", pgError(err))
	}
	for rows.Next() {
		var (
			t time.Time
			o *model.Order
		)
		if err := pgError(rows.Scan(&o.Number, &o.Status, &t)); err != nil {
			return nil, fmt.Errorf("scan rows: %v", err)
		}
		o.UploadedAt = t.Format(time.RFC3339)
		res = append(res, o)
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
	);
	`
	o.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace(debugQuery(q))

	var status bool
	if err := o.s.db.QueryRow(ctx, q, number, user).Scan(&status); err != nil {
		return pgError(err)
	}
	if status {
		return ErrAlreadyRegisteredByUser
	}
	return ErrAlreadyRegisteredByAnotherUser
}

func (o *orderRepository) ChangeStatus(ctx context.Context, m *model.OrderInAccrual) error {
	q := `
		UPDATE
			orders
		SET
			status = $1,
			accrual = $2
		WHERE
			id = $3;
	`
	o.s.logger.Debug(debugQuery(q))

	if _, err := o.s.db.Exec(ctx, q, m.Status, m.Accrual, m.Number); err != nil {
		return fmt.Errorf("exec: %v", pgError(err))
	}
	return nil
}
