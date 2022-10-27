package poller

import (
	"context"
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
		for _, order := range orders {
			l.Trace("push order to queue: %v", order)
			o.queue <- &task{
				ID:     order.Number,
				User:   order.User,
				Status: order.Status,
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
