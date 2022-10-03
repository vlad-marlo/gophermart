package server

import "net/http"

func (s *server) error(w http.ResponseWriter, err error, msg string, status int) {
	http.Error(w, msg, status)
	s.logger.Error(err)
}
