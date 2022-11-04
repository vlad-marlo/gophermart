package poller

import (
	"encoding/json"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"
	"net/http"
)

// GetOrderFromAccrual ...
func (s *Poller) GetOrderFromAccrual(reqID string, number int) (o *model.OrderInAccrual, err error) {

	l := s.logger.WithField("request_id", reqID)
	o = new(model.OrderInAccrual)

	endpoint := fmt.Sprintf("http://%s/api/orders/%d", s.config.AccuralSystemAddress, number)
	l.Tracef("request to %s", endpoint)

	response, err := s.client.R().Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("http get: %w ", err)
	}

	switch response.StatusCode() {
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

	if err := json.Unmarshal(response.Body(), &o); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return o, nil
}
