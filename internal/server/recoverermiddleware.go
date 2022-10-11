package server

import (
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// recoverMiddleware copied github.com/go-chi/chi/v5/middleware.Recoverer(), with my logger
func (s *server) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

				id := middleware.GetReqID(r.Context())
				s.logger.WithField("request_id", id).Tracef("recover covered: %v", rvr)

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
