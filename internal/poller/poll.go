package poller

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type (
	task struct {
		ID, User int
		ReqID    string
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
			l.Errorf("get unprocessed orders: %v", err)
			return
		}
		for _, order := range orders {
			o.queue <- &task{
				ID:    order.Number,
				User:  order.User,
				ReqID: "init poller",
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
					s.logger.WithField("poll", poll).Fatalf("recovered panic in poller: %v", recovered)
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
		s.queue <- &task{num, user, reqID}
	}()
	return nil
}
