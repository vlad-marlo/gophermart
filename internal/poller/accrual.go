package poller

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/vlad-marlo/gophermart/internal/model"
	"net/http"
	"strconv"
	"time"
)

// retryFunc ...
func retryFunc(_ *resty.Client, response *resty.Response) (time.Duration, error) {
	if response.StatusCode() != http.StatusTooManyRequests {
		return 0, nil
	}

	retryAfterValue := response.Header().Get("retry-after")
	if len(retryAfterValue) == 0 {
		return 0, nil
	}

	seconds, err := strconv.Atoi(retryAfterValue)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}

// GetOrderFromAccrual ...
func (s *OrderPoller) GetOrderFromAccrual(number int) (o *model.OrderInAccrual, err error) {

	l := s.logger
	o = new(model.OrderInAccrual)

	endpoint := fmt.Sprintf("%s/api/orders/%d", s.config.AccuralSystemAddress, number)

	r := s.client.NewRequest()
	response, err := r.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("http get: %w ", err)
	}

	tr := r.TraceInfo()
	l.WithFields(map[string]interface{}{
		"status_code": response.StatusCode(),
		"status":      response.Status(),
		"duration":    tr.TotalTime,
		"server_time": tr.ServerTime,
	}).Tracef("request to %s", endpoint)

	switch response.StatusCode() {
	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests
	case http.StatusInternalServerError:
		return nil, ErrInternal
	case http.StatusOK:
		if err := json.Unmarshal(response.Body(), &o); err != nil {
			return nil, fmt.Errorf("json unmarshal: %w", err)
		}
		return o, nil
	default:
		l.Error(fmt.Sprintf("got unexpected status code: %v", response.StatusCode()))
		return nil, ErrUnexpectedStatus
	}
}
