package sqlstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/vlad-marlo/gophermart/internal/model"
)

type withdrawRepository struct {
	s *storage
}

func (r *withdrawRepository) Withdraw(ctx context.Context, w *model.Withdraw) error {
	return nil
}

func (r *withdrawRepository) GetAllByUser(ctx context.Context, user int) (w []*model.Withdraw, err error) {
	q := `
	SELECT 
		order_id, order_sum, processed_at
	FROM 
		withdrawals
	WHERE
		user_id = $1;
	`
	r.s.logger.WithField("request_id", middleware.GetReqID(ctx)).Trace("query: %v", debugQuery(q))

	rows, err := r.s.db.Query(ctx, q, user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoContent
		}
		return nil, fmt.Errorf("query: %v", pgError(err))
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %v", pgError(err))
	}

	for rows.Next() {
		var o *model.Withdraw

		if err := rows.Scan(&o.Order, &o.Sum, &o.ProcessedAt); err != nil {
			return nil, fmt.Errorf("rows scan: %v", pgError(err))
		}

		o.ToRepresentation()
		w = append(w, o)
	}

	if len(w) == 0 {
		return nil, ErrNoContent
	}

	return w, nil
}
