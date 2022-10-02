package server

import (
	"net/http"
	"time"
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
		lrw := newLoggingResponseWriter(w)
		start := time.Now()
		next.ServeHTTP(lrw, r)
		dur := time.Now().Sub(start)
		s.logger.Infof("%s to %s, finished with code %d %s, duration %s", r.Method, r.URL.Path, lrw.statusCode, http.StatusText(lrw.statusCode), dur)
	})
}
