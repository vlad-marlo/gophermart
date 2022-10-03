package server

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type (
	loggingResponseWriter struct {
		http.ResponseWriter
		statusCode int
	}
	// custom type for
	requestIDCtx string
)

var (
	ErrContextDoesntStoreID = errors.New("can't get id from context")
)

const (
	requestIDCtxField requestIDCtx = "id"
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
		lrw := newLoggingResponseWriter(w)

		// request ID
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), requestIDCtxField, id)

		// start check time
		start := time.Now()
		next.ServeHTTP(lrw, r.WithContext(ctx))
		dur := time.Now().Sub(start)

		// log request
		s.logger.Tracef("%s | %s to %s, finished with code %d %s, duration %s", id, r.Method, r.URL.Path, lrw.statusCode, http.StatusText(lrw.statusCode), dur)
	})
}

func GetIDFromContext(ctx context.Context) (string, error) {
	id := ctx.Value(requestIDCtxField)
	if v, ok := id.(string); ok {
		return v, nil
	}
	return "", ErrContextDoesntStoreID
}
