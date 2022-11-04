package poller

import (
	"context"
	"errors"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"
	"time"
)

// pollWork ...
func (s *OrderPoller) pollWork(o *model.OrderInPoll) {
	ctx := context.Background()

	l := s.logger.WithFields(map[string]interface{}{
		"user":  o.User,
		"order": o.Number,
	})

	order, err := s.GetOrderFromAccrual(o.Number)
	if err != nil {
		if errors.Is(err, ErrTooManyRequests) || errors.Is(err, ErrInternal) {
			l.Trace(fmt.Sprintf("pushing order back to queue: got bad status: %v", err))
			time.Sleep(10 * time.Second)
		}
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrNoContent) {
			l.Trace("order was not found in accrual system: deleting order")
		}
		return
	}

	l.Trace(fmt.Sprintf("got order from accrual Order{Status: %s, Accrual: %f, Number: %d}", order.Status, order.Accrual, order.Number))
	switch o.Status {
	case model.StatusProcessing:
		if err := s.store.Order().ChangeStatus(ctx, o.User, order); err != nil {
			l.Warnf("change status: %v", err)
		}
	case "REGISTERED":
	case model.StatusInvalid:
		if err := s.store.Order().ChangeStatus(ctx, o.User, order); err != nil {
			l.Warnf("change status: %v", err)
		}
	case model.StatusProcessed:
		if order.Accrual > 0.0 {
			if err := s.store.Order().ChangeStatusAndIncrementUserBalance(ctx, o.User, order); err != nil {
				return
			}
			l.Trace("successful changed status to processed and incremented user balance")
			return
		} else if err := s.store.Order().ChangeStatus(ctx, o.User, order); err != nil {
			l.Warnf("change status: %v", err)
			return
		}
		l.Trace("successful changed user status")

	default:
		l.Warnf("got unknown status: %s", o.Status)
	}
}
