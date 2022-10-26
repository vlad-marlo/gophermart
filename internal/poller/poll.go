package poller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type (
	task struct {
		ID, User      int
		Status, ReqID string
	}
	OrderPoller struct {
		queue  chan *task
		store  store.Storage
		logger logger.Logger
		config *config.Config
		limit  int
	}
)

func New(l logger.Logger, s store.Storage, limit int) *OrderPoller {
	o := &OrderPoller{
		queue:  make(chan *task, 2*limit),
		store:  s,
		logger: l,
		limit:  limit,
	}
	o.startPolling()
	return o
}

// Close ...
func (s *OrderPoller) Close() {
	close(s.queue)
}

// startPolling ...
func (s *OrderPoller) startPolling() {
	for i := 0; i < s.limit; i++ {
		go func() {
			for t := range s.queue {
				s.pollWork(t)
				if rcvrd := recover(); rcvrd != nil {
					s.logger.Errorf("recovered panic in poller: %v", rcvrd)
				}
			}
		}()
	}
}

// pollWork ...
func (s *OrderPoller) pollWork(t *task) {
	ctx := context.Background()
	l := s.logger.WithFields(map[string]interface{}{
		"request_id": t.ReqID,
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

// GetOrderFromAccrual ...
func (s *OrderPoller) GetOrderFromAccrual(reqID string, number int) (o *model.OrderInAccrual, err error) {
	l := s.logger.WithField("request_id", reqID)
	o = new(model.OrderInAccrual)
	l.Trace("init order in accrual")
	response, err := http.Get(fmt.Sprintf("%s/%d", s.config.AccuralSystemAddres, number))
	if err != nil {
		return nil, fmt.Errorf("http get: %v", err)
	}

	switch response.StatusCode {
	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests
	case http.StatusInternalServerError:
		return nil, ErrInternal
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			s.logger.Warnf("get order form accrual: response body close: %v", err)
		}
	}()

	data, err := io.ReadAll(response.Body)
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, fmt.Errorf("json unmarshal: %v", err)
	}

	return o, nil
}

// Register ...
func (s *OrderPoller) Register(ctx context.Context, user, num int) error {
	err := s.store.Order().Register(ctx, user, num)
	if err != nil {
		s.logger.WithField("request_id", middleware.GetReqID(ctx)).Tracef("err: %v", err)
		return err
	}

	go func() {
		s.queue <- &task{num, user, model.StatusNew, middleware.GetReqID(ctx)}
	}()
	return nil
}
