package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
)

type (
	// add possibility to log status code after handling request
	loggingRW struct {
		http.ResponseWriter
		statusCode int
	}
	// universal interface for logger middleware
)

// newLoggingRW ...
func newLoggingRW(w http.ResponseWriter) *loggingRW {
	return &loggingRW{w, http.StatusOK}
}

// WriteHeader ...
func (l *loggingRW) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

// LogRequest ...
func LogRequest(logger logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := middleware.GetReqID(r.Context())
			lrw := newLoggingRW(w)
			// start check time
			start := time.Now()
			next.ServeHTTP(lrw, r)
			dur := time.Now().Sub(start)

			// log request
			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"url":        r.URL.Path,
				"duration":   fmt.Sprint(dur),
				"code":       lrw.statusCode,
				"request_id": id,
			}).Trace(http.StatusText(lrw.statusCode))
		})
	}
}
