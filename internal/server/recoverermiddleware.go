package server

import (
	"net/http"
)

// recoverMiddleware copied github.com/go-chi/chi/v5/middleware.Recoverer(), with my logger
func (s *server) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

				id, err := GetIDFromContext(r.Context())
				if err != nil {
					s.logger.Debugf("recover covered: %v", rvr)
					return
				}
				s.logger.Tracef("%s | recover covered: %v", id, rvr)

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
