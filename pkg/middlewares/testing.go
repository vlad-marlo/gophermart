package middlewares

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

// TestServer ...
func TestServer(t *testing.T) chi.Router {
	t.Helper()

	r := chi.NewMux()
	log := logrus.New()
	log.Out = io.Discard
	l := logger.GetLoggerByEntry(logrus.NewEntry(log))
	r.Use(Recover(l))
	r.Get("/", func(_ http.ResponseWriter, _ *http.Request) {
		panic("something went wrong")
	})
	return r
}
