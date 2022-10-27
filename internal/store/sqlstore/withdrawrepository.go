package sqlstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type withdrawRepository struct {
	s *storage
}

func (r *withdrawRepository) Migrate(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS withdrawals(
			id BIGSERIAL UNIQUE PRIMARY KEY,
			processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			user_id BIGINT,
			order_id BIGINT,
			order_sum INT NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`

	if _, err := r.s.db.Exec(ctx, q); err != nil {
		return fmt.Errorf("query: %v", pgError(err))
	}
	return nil
}

func (r *withdrawRepository) Withdraw(ctx context.Context, user int, w *model.Withdraw) error {
	var bal int
	qGetBal := `
	SELECT
		balance::numeric::int
	FROM
		users
	WHERE
		id = $1;
	`
	qWithdraw := `
	UPDATE
		users
	SET
		balance = balance - $1::numeric::money
	WHERE
		id = $2;
	`

	qInsertWithdrawal := `
	INSERT INTO
		withdrawals(
		    user_id,
		    order_id,
		    order_sum
		)
	VALUES ($1, $2, $3);	                                                                        
	`

	fields := map[string]interface{}{
		"request_id": middleware.GetReqID(ctx),
	}
	l := r.s.logger.WithFields(fields)

	tx, err := r.s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx begin: %v", pgError(err))
	}
	l.Trace("tx started")

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			l.Fatalf("unable to update drivers: %v", err)
			return
		}
		l.Trace("tx rollbacked")
	}()

	if err := tx.QueryRow(ctx, qGetBal, user).Scan(&bal); err != nil {
		return fmt.Errorf("get balance: %v", pgError(err))
	}

	l.Trace("successful get balance")
	if bal < w.Sum {
		return store.ErrPaymentRequired
	}
	l.Trace("balance is ok")

	if _, err := tx.Exec(ctx, qWithdraw, w.Sum, user); err != nil {
		return fmt.Errorf("withdraw balance: %v", pgError(err))
	}
	l.Trace("balance withdrawn successful")

	if _, err := tx.Exec(ctx, qInsertWithdrawal, user, w.Order, w.Sum); err != nil {
		l.WithField("sql", debugQuery(qInsertWithdrawal)).Tracef("%v", pgError(err))
		return fmt.Errorf("update withdraw: %v", pgError(err))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("update drivers: %v", pgError(err))
	}
	return nil
}

func (r *withdrawRepository) GetAllByUser(ctx context.Context, user int) (res []*model.Withdraw, err error) {
	q := `
	SELECT 
		order_id, order_sum, processed_at
	FROM 
		withdrawals
	WHERE
		user_id = $1
	ORDER BY processed_at;
	`
	r.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace("query: %v", debugQuery(q))

	rows, err := r.s.db.Query(ctx, q, user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNoContent
		}
		return nil, fmt.Errorf("query: %v", pgError(err))
	}

	defer rows.Close()

	for rows.Next() {
		o := new(model.Withdraw)

		if err := rows.Scan(&o.Order, &o.Sum, &o.ProcessedAt); err != nil {
			return nil, fmt.Errorf("rows scan: %v", pgError(err))
		}

		o.ToRepresentation()
		res = append(res, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %v", pgError(err))
	}

	if len(res) == 0 {
		return nil, store.ErrNoContent
	}

	return res, nil
}
