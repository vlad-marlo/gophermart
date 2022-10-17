package poller

import (
	"sync"
	"time"

	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type (
	OrderPoller struct {
		queue  []chan int
		qSize  int
		qLimit int
		s      store.Storage
		l      logger.Logger
	}
)

func New(l logger.Logger, s store.Storage, limit int) *OrderPoller {
	o := &OrderPoller{
		queue:  make([]chan int, limit),
		qSize:  0,
		qLimit: limit,
		s:      s,
		l:      l,
	}
	go func(o *OrderPoller) {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
			default:
				return
			}
		}
	}(o)

	return o
}

// deque FanIn flushing buffer
func (t *OrderPoller) deque() {
	wg := &sync.WaitGroup{}
	for _, ch := range t.queue {
		wg.Add(1)
		go func(ch chan int) {
			defer wg.Done()
		}(ch)
		wg.Wait()
	}
}
