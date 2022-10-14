package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetReqID(r.Context())
		lrw := newLoggingResponseWriter(w)
		// start check time
		start := time.Now()
		next.ServeHTTP(lrw, r)
		dur := time.Now().Sub(start)

		// log request
		s.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"url":        r.URL.Path,
			"duration":   fmt.Sprint(dur),
			"code":       lrw.statusCode,
			"request_id": id,
		}).Trace(http.StatusText(lrw.statusCode))
	})
}
