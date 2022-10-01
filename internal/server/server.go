package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/storage"
)

type server struct {
	chi.Router
	store storage.Storage
	// don't sure that this is necessary
	config *config.Config
}

func Start(store storage.Storage, config *config.Config) error {
	s := &server{
		store:  store,
		config: config,
		Router: chi.NewMux(),
	}
	s.configureMiddlewares()
	return http.ListenAndServe(s.config.BindAddr, s.Router)
}

func (s *server) configureMiddlewares() {
	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)
	s.Use(middleware.Compress(5, "text/html", "text/plain", "application/json"))
}
