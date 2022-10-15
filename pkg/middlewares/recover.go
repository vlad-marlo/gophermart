package middlewares

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

func Recover(logger logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

					id := middleware.GetReqID(r.Context())
					logger.WithField("request_id", id).Tracef("recover covered: %v", rvr)

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
