package middlewares

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

// Recover ...
func Recover(logger logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

					id := middleware.GetReqID(r.Context())
					logger.WithField("request_id", id).Error(fmt.Sprintf("recover covered: %v", rvr))

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
