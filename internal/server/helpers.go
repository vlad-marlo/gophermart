package server

import "net/http"

func (s *server) error(w http.ResponseWriter, err error, status int) {
	http.Error(w, err.Error(), status)
	s.logger.Error(err)
}
