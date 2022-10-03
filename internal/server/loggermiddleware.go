package server

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/pkg/utils"
	"net/http"
	"time"
)

type (
	loggingResponseWriter struct {
		http.ResponseWriter
		statusCode int
	}
)

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		lrw := newLoggingResponseWriter(w)
		// start check time
		start := time.Now()
		next.ServeHTTP(lrw, r.WithContext(utils.GetCtxWithID(r.Context(), id)))
		dur := time.Now().Sub(start)

		// log request
		s.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"url":        r.URL.Path,
			"duration":   dur,
			"code":       lrw.statusCode,
			"request_id": id,
		}).Trace(http.StatusText(lrw.statusCode))
	})
}
