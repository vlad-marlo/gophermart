package poller

import (
	"context"
	"errors"
	"github.com/vlad-marlo/gophermart/internal/model"
)

// pollWork ...
func (s *OrderPoller) pollWork(poller int, t *task) {
	ctx := context.Background()
	l := s.logger.WithFields(map[string]interface{}{
		"request_id": t.ReqID,
		"poll":       poller,
	})

	l.Trace("get order from accrual")
	o, err := s.GetOrderFromAccrual(t.ReqID, t.ID)
	if err != nil {
		l.Debugf("pool work: %v", err)

		if errors.Is(err, ErrInternal) || errors.Is(err, ErrTooManyRequests) {
			l.Trace("sending task to queue")
			s.queue <- t
			return
		}
		return
	}

	l.Trace("get status")
	switch o.Status {

	case "PROCESSING":
		o.Status = model.StatusProcessing
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			l.Debugf("change status: %v", err)
		}
		l.Trace("status changed; sending updated task to queue")
		s.queue <- &task{t.ID, t.User, model.StatusProcessing, t.ReqID}

	case "REGISTERED":
		l.Trace("only registered")
		s.queue <- t

	case "INVALID":
		l.Trace("invalid status stop polling")
		o.Status = model.StatusInvalid
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			l.Tracef("change status: %v", err)
		}

	case "PROCESSED":
		o.Status = model.StatusProcessed
		if o.Accrual > 0 {
			if err := s.store.User().IncrementBalance(ctx, t.User, o.Accrual); err != nil {
				l.Tracef("increment user balance: %v", err)
			}
			l.Trace("successful incremented balance")
		}
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			l.Tracef("change status: %v", err)
		}

	default:
		l.Warnf("got unknown status: %v", o.Status)
	}
}
