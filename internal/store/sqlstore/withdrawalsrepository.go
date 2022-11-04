package sqlstore

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/luhn"
)

type withdrawRepository struct {
	s *storage
}

func (r *withdrawRepository) Migrate(ctx context.Context) error {
	q := debugQuery(`CREATE TABLE IF NOT EXISTS withdrawals(
			id BIGSERIAL UNIQUE PRIMARY KEY,
			processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			user_id BIGINT,
			order_id BIGINT UNIQUE,
			order_sum DOUBLE PRECISION DEFAULT 0::DOUBLE PRECISION,

			FOREIGN KEY (user_id) REFERENCES users(id)
		);
-- 			FOREIGN KEY (order_id) REFERENCES orders(id),`)

	if _, err := r.s.db.Exec(ctx, q); err != nil {
		return pgError("query: %w", err)
	}
	return nil
}

func (r *withdrawRepository) Withdraw(ctx context.Context, user int, w *model.Withdraw) error {
	if !luhn.Valid(w.Order) {
		return store.ErrIncorrectData
	}
	var bal float64
	qGetBal := debugQuery(`
	SELECT
		balance::FLOAT8
	FROM
		users
	WHERE
		id = $1;
	`)
	qWithdraw := debugQuery(`
	UPDATE
		users
	SET
		balance = balance - $1::DOUBLE PRECISION
	WHERE
		id = $2;
	`)

	//qOrderRegisteredByUser := debugQuery(`
	//SELECT EXISTS(
	//	SELECT
	//	    *
	//	FROM
	//	    orders
	//    WHERE
	//        user_id = $1 AND id = $2
	//);
	//`)

	qInsertWithdrawal := debugQuery(`
	INSERT INTO
		withdrawals(
		    user_id,
		    order_id,
		    order_sum
		)
	VALUES ($1, $2, $3);	                                                                        
	`)

	tx, err := r.s.db.Begin(ctx)
	if err != nil {
		return pgError("tx begin: %w", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.s.logger.WithFields(map[string]interface{}{
				"request_id": middleware.GetReqID(ctx),
			}).Fatal(pgError("unable to update drivers: %w", err))
			return
		}
	}()

	if err := tx.QueryRow(ctx, qGetBal, user).Scan(&bal); err != nil {
		return pgError("get balance: %w", err)
	}

	if bal < w.Sum {
		return store.ErrPaymentRequired
	}

	//var ok bool
	//if err := tx.QueryRow(ctx, qOrderRegisteredByUser, user, w.Order).Scan(&ok); err != nil {
	//	return pgError("check is order registered by user or not", err)
	//}
	//if !ok {
	//	return store.ErrAlreadyRegisteredByAnotherUser
	//}

	if _, err := tx.Exec(ctx, qWithdraw, w.Sum, user); err != nil {
		return pgError("withdraw balance: %w", err)
	}

	if _, err := tx.Exec(ctx, qInsertWithdrawal, user, w.Order, w.Sum); err != nil {
		return pgError("update withdraw: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return pgError("update drivers: %w", err)
	}
	return nil
}

func (r *withdrawRepository) GetAllByUser(ctx context.Context, user int) (res []*model.Withdraw, err error) {
	q := debugQuery(`
	SELECT 
		order_id, order_sum::FLOAT8, processed_at
	FROM 
		withdrawals
	WHERE
		user_id = $1
	ORDER BY processed_at;
	`)
	rows, err := r.s.db.Query(ctx, q, user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNoContent
		}
		return nil, pgError("query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		o := new(model.Withdraw)

		if err := rows.Scan(&o.Order, &o.Sum, &o.ProcessedAt); err != nil {
			return nil, pgError("rows scan: %w", err)
		}

		o.ToRepresentation()
		res = append(res, o)
	}

	if err := rows.Err(); err != nil {
		return nil, pgError("rows err: %w", err)
	}

	if len(res) == 0 {
		return nil, store.ErrNoContent
	}

	return res, nil
}
