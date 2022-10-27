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

	o, err := s.GetOrderFromAccrual(t.ReqID, t.ID)
	if err != nil {

		if errors.Is(err, ErrInternal) || errors.Is(err, ErrTooManyRequests) {
			s.queue <- t
			return
		}
		return
	}

	switch o.Status {

	case "PROCESSING":
		if o.Status == model.StatusProcessing {
			s.queue <- &task{t.ID, t.User, model.StatusProcessing, t.ReqID}
			return
		}

		o.Status = model.StatusProcessing
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			l.Warnf("change status: %v", err)
		}
		s.queue <- &task{t.ID, t.User, model.StatusProcessing, t.ReqID}

	case "REGISTERED":
		s.queue <- t

	case model.StatusInvalid:
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			l.Warnf("change status: %v", err)
		}
	case model.StatusProcessed:
		if o.Accrual > 0.0 {
			if err := s.store.User().IncrementBalance(ctx, t.User, o.Accrual); err != nil {
				l.Tracef("increment user balance: %v", err)
			}
			l.Trace("successful incremented balance")
		}
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			l.Warnf("change status: %v", err)
		}

	default:
		l.Warnf("got unknown status: %v", o.Status)
	}
}
