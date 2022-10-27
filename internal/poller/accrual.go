package poller

import (
	"encoding/json"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"
	"io"
	"net/http"
)

// GetOrderFromAccrual ...
func (s *OrderPoller) GetOrderFromAccrual(reqID string, number int) (o *model.OrderInAccrual, err error) {
	l := s.logger.WithField("request_id", reqID)
	o = new(model.OrderInAccrual)

	endpoint := fmt.Sprintf("http://%s/api/orders/%d", s.config.AccuralSystemAddress, number)
	l.Tracef("request to %s", endpoint)

	response, err := http.Get(endpoint)
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
