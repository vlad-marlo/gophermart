package poller

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"time"
)

type (
	OrderPoller struct {
		queue  chan struct{}
		store  store.Storage
		logger logger.Logger
		config config.Config
		client *resty.Client
	}
)

func New(l logger.Logger, s store.Storage, cfg config.Config, interval time.Duration) *OrderPoller {
	p := &OrderPoller{
		queue:  make(chan struct{}),
		store:  s,
		logger: l,
		config: cfg,
		client: resty.New().SetRetryAfter(retryFunc).SetRetryCount(3),
	}

	go func() {
		t := time.NewTicker(interval)
		for {
			select {
			case <-t.C:
				orders, err := p.store.Order().GetUnprocessedOrders(context.Background())
				if err != nil && !errors.Is(err, store.ErrNoContent) {
					l.Error(fmt.Sprintf("get unprocessed orders: %v", err))
					continue
				}
				for _, order := range orders {
					p.pollWork(order)
				}
			case <-p.queue:
				l.Trace("graceful closed poller")
				return
			}
		}
	}()
	return p
}

// Close ...
func (s *OrderPoller) Close() {
	close(s.queue)
}
