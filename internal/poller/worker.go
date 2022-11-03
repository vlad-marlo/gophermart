package poller

import (
	"context"
	"errors"
	"github.com/vlad-marlo/gophermart/internal/model"
	"time"
)

// pollWork ...
func (s *Poller) pollWork(poller int, t *task) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := s.logger.WithFields(map[string]interface{}{
		"request_id": t.ReqID,
		"poll":       poller,
	})

	o, err := s.GetOrderFromAccrual(t.ReqID, t.ID)
	if err != nil {
		if errors.Is(err, ErrTooManyRequests) {
			time.Sleep(10 * time.Second)
			s.queue <- t
		}
		return
	}

	switch o.Status {

	case model.StatusProcessing:
		o.Status = model.StatusProcessing
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
			}
			return
		}
		if err := s.store.Order().ChangeStatus(ctx, t.User, o); err != nil {
			s.queue <- &task{t.ID, t.User, t.ReqID}
			l.Warnf("change status: %v", err)
		}

	default:
		l.Warnf("got unknown status: %s", o.Status)
	}
}
