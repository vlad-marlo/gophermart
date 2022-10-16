package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/model"
	"io"
	"net/http"

	"github.com/vlad-marlo/gophermart/pkg/logger"
	"github.com/vlad-marlo/gophermart/pkg/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/store"
)

type server struct {
	chi.Router
	store  store.Storage
	logger logger.Logger
	// don't sure that config is necessary in server struct
	config *config.Config
}

// Start ...
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

// configureMiddlewares ...
func (s *server) configureMiddlewares() {
	s.Use(middleware.RequestID)
	s.Use(middlewares.LogRequest(s.logger))

	s.Use(middlewares.Recover(s.logger))
	s.Use(middleware.Compress(5, "text/html", "text/plain", "application/json"))
}

// configureRoutes ...
func (s *server) configureRoutes() {
	s.Route("/api/user", func(r chi.Router) {
		r.Post("/register", s.handleAuthRegister())
		r.Post("/login", s.handleAuthLogin())
		// endpoints for authorized users only
		r.With(s.CheckAuthMiddleware).Route("/", func(r chi.Router) {
			r.Post("/orders", s.handleOrdersPost())
			r.Get("/orders", s.handleOrdersGet())
			r.Get("/balance", s.handleBalanceGet())
			r.Post("/balance/withdraw", s.handleBalanceWithdrawPost())
			r.Get("/balance/withdrawals", s.handleGetAllWithdraws())
		})
	})
}

func (s *server) GetOrderFromAccrual(number string) (o *model.OrderInAccrual, err error) {
	response, err := http.Get(fmt.Sprintf("%s/%s", s.config.AccuralSystemAddres, number))
	if err != nil {
		return nil, fmt.Errorf("http get: %v", err)
	}

	switch response.StatusCode {
	case http.StatusTooManyRequests:
		return nil, errors.New("to many request")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("%v", InternalErrMsg)
	}

	defer func() {
		_ = response.Body.Close()
	}()

	data, err := io.ReadAll(response.Body)
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, fmt.Errorf("json unmarshal: %v", err)
	}

	return o, nil
}
