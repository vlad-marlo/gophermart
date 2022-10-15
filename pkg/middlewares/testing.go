package middlewares

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

func TestServer(t *testing.T) chi.Router {
	t.Helper()

	r := chi.NewMux()
	l := logger.GetLogger()
	r.Use(Recover(l))
	r.Get("/", func(_ http.ResponseWriter, _ *http.Request) {
		panic("something went wrong")
	})
	return r
}
