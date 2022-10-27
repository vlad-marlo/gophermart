package sqlstore

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
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
		return pgError("query: %v", err)
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

	tx, err := r.s.db.Begin(ctx)
	if err != nil {
		return pgError("tx begin: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.s.logger.WithFields(map[string]interface{}{
				"request_id": middleware.GetReqID(ctx),
			}).Fatal(pgError("unable to update drivers: %v", err))
			return
		}
	}()

	if err := tx.QueryRow(ctx, qGetBal, user).Scan(&bal); err != nil {
		return pgError("get balance: %v", err)
	}

	if bal < w.Sum {
		return store.ErrPaymentRequired
	}

	if _, err := tx.Exec(ctx, qWithdraw, w.Sum, user); err != nil {
		return pgError("withdraw balance: %v", err)
	}

	if _, err := tx.Exec(ctx, qInsertWithdrawal, user, w.Order, w.Sum); err != nil {
		return pgError("update withdraw: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return pgError("update drivers: %v", err)
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
		return nil, pgError("query: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		o := new(model.Withdraw)

		if err := rows.Scan(&o.Order, &o.Sum, &o.ProcessedAt); err != nil {
			return nil, pgError("rows scan: %v", err)
		}

		o.ToRepresentation()
		res = append(res, o)
	}

	if err := rows.Err(); err != nil {
		return nil, pgError("rows err: %v", err)
	}

	if len(res) == 0 {
		return nil, store.ErrNoContent
	}

	return res, nil
}
