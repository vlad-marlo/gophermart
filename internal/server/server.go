package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type server struct {
	chi.Router
	store  store.Storage
	logger *logrus.Logger
	// don't sure that this is necessary
	config *config.Config
}

func Start(logger *logrus.Logger, store store.Storage, config *config.Config) error {
	s := &server{
		store:  store,
		config: config,
		Router: chi.NewMux(),
		logger: logger,
	}
	s.configureMiddlewares()
	return http.ListenAndServe(s.config.BindAddr, s.Router)
}

func (s *server) configureMiddlewares() {
	s.Use(s.loggerMiddleware)
	s.Use(middleware.Recoverer)
	s.Use(middleware.Compress(5, "text/html", "text/plain", "application/json"))
}

func (s *server) confugureRoutes() {}
