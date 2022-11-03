package poller

import (
	"encoding/json"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"
	"io"
	"net/http"
)

// GetOrderFromAccrual ...
func (s *Poller) GetOrderFromAccrual(reqID string, number int) (o *model.OrderInAccrual, err error) {
	l := s.logger.WithField("request_id", reqID)
	o = new(model.OrderInAccrual)

	endpoint := fmt.Sprintf("http://%s/api/orders/%d", s.config.AccuralSystemAddress, number)
	l.Tracef("request to %s", endpoint)

	response, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("http get: %w ", err)
	}
	defer func() {
		if _, err := io.Copy(io.Discard, response.Body); err != nil {
			l.Warnf("io copy: %v", err)
		}
		if err := response.Body.Close(); err != nil {
			l.Warnf("get order form accrual: response body close: %v", err)
		}
	}()

	switch response.StatusCode {
	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests
	case http.StatusInternalServerError:
		return nil, ErrInternal
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusNoContent:
		return nil, ErrNoContent
	default:
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if err := json.Unmarshal(data, &o); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return o, nil
}
