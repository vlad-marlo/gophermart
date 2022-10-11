package server

import "net/http"

func (s *server) error(w http.ResponseWriter, err error, msg, id string, status int) {
	http.Error(w, msg, status)
	s.logger.WithField(RequestIDLoggerField, id).Error(err)
}
