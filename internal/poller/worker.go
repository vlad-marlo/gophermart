package poller

import (
	"context"
	"errors"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"
	"time"
)

// pollWork ...
func (s *Poller) pollWork(poller int, t *task) {
	ctx := context.Background()

	l := s.logger.WithFields(map[string]interface{}{
		"request_id": t.ReqID,
		"poll":       poller,
	})

	o, err := s.GetOrderFromAccrual(t.ReqID, t.ID)
	if err != nil {
		if errors.Is(err, ErrTooManyRequests) || errors.Is(err, ErrInternal) {
			time.Sleep(10 * time.Second)
			s.queue <- t
		}
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrNoContent) {
			l.Trace("order was not found in accrual system: deleting order")
			if err := s.store.Order().Delete(ctx, t.User, t.ID); err != nil {
				s.logger.Error(fmt.Sprintf("delete order: %v", err))
			}
		}
		return
	}

	l.Trace(fmt.Sprintf("got order from accrual Order{Status: %s, Accrual: %f, Number: %d}", o.Status, o.Accrual, o.Number))
	switch o.Status {
	case model.StatusProcessing:
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			l.Warnf("change status: %v", err)
		}
		s.queue <- &task{t.ID, t.User, t.ReqID}
	case "REGISTERED":
		s.queue <- t
	case model.StatusInvalid:
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			s.queue <- &task{t.ID, t.User, t.ReqID}
			l.Warnf("change status: %v", err)
		}
	case model.StatusProcessed:
		if o.Accrual > 0.0 {
			if err := s.store.Order().ChangeStatusAndIncrementUserBalance(ctx, t.User, o); err != nil {
				s.queue <- &task{t.ID, t.User, t.ReqID}
				return
			}
			l.Trace("successful changed status to processed and incremented user balance")
			return
		} else if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			s.queue <- &task{t.ID, t.User, t.ReqID}
			l.Warnf("change status: %v", err)
			return
		}
		l.Trace("successful changed user status")

	default:
		l.Warnf("got unknown status: %s", o.Status)
	}
}
