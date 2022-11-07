package poller

import (
	"context"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"
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
		return
	}

	l.Trace(fmt.Sprintf("got order from accrual Order{Status: %s, Accrual: %f, Number: %d}", order.Status, order.Accrual, order.Number))
	switch order.Status {
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
