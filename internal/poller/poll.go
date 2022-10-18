package poller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type (
	task struct {
		ID, User int
		Status   string
	}
	OrderPoller struct {
		queue  chan task
		store  store.Storage
		logger logger.Logger
		config *config.Config
		limit  int
	}
)

var poll uint64 = 0

func New(l logger.Logger, s store.Storage, limit int) *OrderPoller {
	o := &OrderPoller{
		queue:  make(chan task, 2*limit),
		store:  s,
		logger: l,
		limit:  limit,
	}
	go o.startPolling()

	return o
}

// deque FanIn flushing buffer
func (s *OrderPoller) Close() {
	close(s.queue)
}

// startPolling ...
func (s *OrderPoller) startPolling() {
	for i := 0; i < s.limit; i++ {
		go func() {
			for t := range s.queue {
				s.pollWork(t)
			}
		}()
	}
}

// pollWork ...
func (s *OrderPoller) pollWork(t task) {
	o, err := s.GetOrderFromAccrual(t.ID)
	if err != nil {
		s.logger.Debugf("pool work: %v", err)
		if errors.Is(err, ErrInternal) || errors.Is(err, ErrTooManyRequests) {
			s.queue <- t
			return
		}
		return
	}
	switch o.Status {
	case "PROCESSING":
		o.Status = model.StatusProcessing
		s.store.Order().ChangeStatus(context.TODO(), o)
		s.queue <- task{t.ID, t.User, model.StatusProcessing}
	case "REGISTERED":
		s.queue <- t
	case "INVALID":
		o.Status = model.StatusInvalid
		s.store.Order().ChangeStatus(context.TODO(), o)
	case "PROCESSED":
		o.Status = model.StatusProcessed
		// TODO добавить пользователям баллов тута
		s.store.Order().ChangeStatus(context.TODO(), o)
	default:
		s.logger.Warnf("got unknown status: %v", o.Status)
	}
}

// GetOrderFromAccrual ...
func (s *OrderPoller) GetOrderFromAccrual(number int) (o *model.OrderInAccrual, err error) {
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
			s.logger.Warn("get order form accrual: response body close: %v", err)
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
	go func() {
		s.queue <- task{num, user, model.StatusNew}
	}()
	return err
}
