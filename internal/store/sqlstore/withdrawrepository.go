package sqlstore

import (
	"context"
	"errors"
	"fmt"
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
		return sqlErr("query: %v", err, q)
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
		return fmt.Errorf("tx begin: %v", pgError(err))
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.s.logger.WithFields(map[string]interface{}{
				"request_id": middleware.GetReqID(ctx),
			}).Fatalf("unable to update drivers: %v", pgError(err))
			return
		}
	}()

	if err := tx.QueryRow(ctx, qGetBal, user).Scan(&bal); err != nil {
		return sqlErr("get balance: %v", err, qGetBal)
	}

	if bal < w.Sum {
		return store.ErrPaymentRequired
	}

	if _, err := tx.Exec(ctx, qWithdraw, w.Sum, user); err != nil {
		return sqlErr("withdraw balance: %v", err, qWithdraw)
	}

	if _, err := tx.Exec(ctx, qInsertWithdrawal, user, w.Order, w.Sum); err != nil {
		return sqlErr("update withdraw: %v", err, qInsertWithdrawal)
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
		return nil, sqlErr("query: %v", err, q)
	}

	defer rows.Close()

	for rows.Next() {
		o := new(model.Withdraw)

		if err := rows.Scan(&o.Order, &o.Sum, &o.ProcessedAt); err != nil {
			return nil, sqlErr("rows scan: %v", err, q)
		}

		o.ToRepresentation()
		res = append(res, o)
	}

	if err := rows.Err(); err != nil {
		return nil, sqlErr("rows err: %v", err, q)
	}

	if len(res) == 0 {
		return nil, store.ErrNoContent
	}

	return res, nil
}
