package poller

import (
	"context"
	"fmt"
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
	OrderPoller interface {
		Register(ctx context.Context, user, num int) error
		Close()
	}
	Poller struct {
		queue  chan *task
		store  store.Storage
		logger logger.Logger
		config *config.Config
		limit  int
	}
)

func New(l logger.Logger, s store.Storage, cfg *config.Config, limit int) OrderPoller {
	p := &Poller{
		queue:  make(chan *task, 20*limit),
		store:  s,
		logger: l,
		limit:  limit,
		config: cfg,
	}
	p.startPolling()
	go func() {
		orders, err := p.store.Order().GetUnprocessedOrders(context.Background())
		if err != nil {
			l.Errorf("get unprocessed orders: %v", err)
			return
		}
		for _, order := range orders {
			p.queue <- &task{
				ID:    order.Number,
				User:  order.User,
				ReqID: "init poller",
			}
		}
	}()
	return p
}

// Close ...
func (s *Poller) Close() {
	close(s.queue)
}

func (s *Poller) poll(poll int) {
	defer func() {
		if recovered := recover(); recovered != nil {
			s.logger.WithField("poll", poll).Error(fmt.Sprintf("recovered panic in poller: %v", recovered))
			go s.poll(poll)
		}
	}()
	for t := range s.queue {
		s.pollWork(poll, t)
	}
}

// startPolling ...
func (s *Poller) startPolling() {
	for i := 0; i < s.limit; i++ {
		go s.poll(i)
	}
}

// Register ...
func (s *Poller) Register(ctx context.Context, user, num int) error {
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
