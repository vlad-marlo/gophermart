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

func New(l logger.Logger, s store.Storage, cfg *config.Config, limit int) *OrderPoller {
	o := &OrderPoller{
		queue:  make(chan *task, 2*limit),
		store:  s,
		logger: l,
		limit:  limit,
		config: cfg,
	}
	o.startPolling()
	go func() {
		orders, err := o.store.Order().GetUnprocessedOrders(context.Background())
		if err != nil {
			l.Trace("get unprocessed orders: %v", err)
			return
		}
		if len(orders) == 0 {
			l.Trace("no unprocessed orders")
		}
		for _, ordr := range orders {
			l.Trace("push order to queue: %v", ordr)
			o.queue <- &task{
				ID:     ordr.Number,
				User:   ordr.User,
				Status: ordr.Status,
				ReqID:  "init poller",
			}
		}
	}()
	return o
}

// Close ...
func (s *OrderPoller) Close() {
	close(s.queue)
}

// startPolling ...
func (s *OrderPoller) startPolling() {
	for i := 0; i < s.limit; i++ {
		go func(poll int) {
			for t := range s.queue {
				s.pollWork(poll, t)
				if recovered := recover(); recovered != nil {
					s.logger.WithField("poll", poll).Errorf("recovered panic in poller: %v", recovered)
				}
			}
		}(i)
	}
}

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

// GetOrderFromAccrual ...
func (s *OrderPoller) GetOrderFromAccrual(reqID string, number int) (o *model.OrderInAccrual, err error) {
	l := s.logger.WithField("request_id", reqID)
	o = new(model.OrderInAccrual)

	response, err := http.Get(fmt.Sprintf("http://%s/api/orders/%d", s.config.AccuralSystemAddres, number))
	if err != nil {
		return nil, fmt.Errorf("http get: %v", err)
	}

	switch response.StatusCode {
	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests
	case http.StatusInternalServerError:
		return nil, ErrInternal
	case http.StatusNotFound:
		return nil, ErrInternal
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			l.Warnf("get order form accrual: response body close: %v", err)
		}
	}()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %v", err)
	}
	l.Trace(string(data), response.StatusCode)
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, fmt.Errorf("json unmarshal: %v", err)
	}

	return o, nil
}

// Register ...
func (s *OrderPoller) Register(ctx context.Context, user, num int) error {
	err := s.store.Order().Register(ctx, user, num)
	if err != nil {
		return err
	}

	go func() {
		reqID := middleware.GetReqID(ctx)
		s.logger.WithField("request_id", reqID).Trace("pushing task to queue")
		s.queue <- &task{num, user, model.StatusNew, reqID}
	}()
	return nil
}
