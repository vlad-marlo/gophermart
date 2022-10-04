package server

import (
	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type server struct {
	chi.Router
	store  store.Storage
	logger logger.Logger
	// don't sure that this is necessary
	config *config.Config
}

func Start(l logger.Logger, store store.Storage, config *config.Config) error {
	s := &server{
		store:  store,
		config: config,
		Router: chi.NewMux(),
		logger: l,
	}
	s.configureMiddlewares()
	s.configureRoutes()
	return http.ListenAndServe(s.config.BindAddr, s.Router)
}

func (s *server) configureMiddlewares() {

	s.Use(middleware.RequestID)
	s.Use(s.logRequest)

	s.Use(middleware.Recoverer)
	s.Use(middleware.Compress(5, "text/html", "text/plain", "application/json"))
}

func (s *server) configureRoutes() {
	s.Route("/api/user", func(r chi.Router) {
		r.Post("/register", s.handleAuthRegister())
		r.Post("/login", s.handleAuthLogin())
	})
}
