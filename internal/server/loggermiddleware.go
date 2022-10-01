package server

import (
	"net/http"
	"time"
)

type loggerRespWriter struct {
	code int
	http.ResponseWriter
}

func (l loggerRespWriter) WriteHeader(code int) {
	l.code = code
	l.WriteHeader(code)
}

func (s *server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := loggerRespWriter{200, w}
		start := time.Now()
		next.ServeHTTP(wr, r)
		dur := time.Now().Sub(start)
		s.logger.Printf("%s finished with code %d by %d", r.URL.String(), wr.code, dur)
	})
}
