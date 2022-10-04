package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"
	"io"
	"net/http"

	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

const (
	BadRequestMsg  = "bad request"
	InternalErrMsg = "internal server error"
)

func (s *server) handleAuthRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u *model.User
		id := middleware.GetReqID(r.Context())

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.logger.Tracef("%v | handle auth register: close body: %v", id, err)
			}
		}()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("auth: register: read data from request: %v", err), InternalErrMsg, id, http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &u); err != nil {
			s.error(w, fmt.Errorf("json unmarshal: %v", err), BadRequestMsg, id, http.StatusBadRequest)
			return
		}

		if err := u.BeforeCreate(); err != nil {
			s.error(w, fmt.Errorf("user: before create: %v", err), InternalErrMsg, id, http.StatusInternalServerError)
			return
		}

		if err := s.store.User().Create(r.Context(), u); err != nil {
			if errors.Is(err, sqlstore.ErrLoginAlreadyInUse) {
				s.error(w, fmt.Errorf("auth register: create user: %v", err), err.Error(), id, http.StatusConflict)
				return
			}
			s.error(w, fmt.Errorf("auth register: create user: %v", err), InternalErrMsg, id, http.StatusInternalServerError)
			return
		}

		s.logger.WithFields(logrus.Fields{
			UserIDLoggerField: u.ID,
		}).Trace("successful authenticated")

		s.authenticate(w, u.ID)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) handleAuthLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *model.User
		id := middleware.GetReqID(r.Context())

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.logger.WithField(RequestIDLoggerField, id).Tracef("%v | handle auth login: close body: %v", id, err)
			}
		}()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("auth: register: read data from request: %v", err), "internal server error", id, http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &req); err != nil {
			s.error(w, fmt.Errorf("login: uncorrect request data: %v", err), "bad request", id, http.StatusBadRequest)
			return
		}

		user, err := s.store.User().GetByLogin(r.Context(), req.Login)
		if err != nil {
			s.error(w, fmt.Errorf("login: unauthorized: %v", err), sqlstore.ErrUncorrectLoginData.Error(), id, http.StatusUnauthorized)
			return
		}
		if err := user.ComparePassword(req.Password); err != nil {
			s.error(w, fmt.Errorf("login: compare pass: unauthorized: %v", err), sqlstore.ErrUncorrectLoginData.Error(), id, http.StatusUnauthorized)
			return
		}

		s.logger.WithFields(logrus.Fields{
			UserIDLoggerField: user.ID,
		}).Trace("successful authenticated")

		s.authenticate(w, user.ID)
		w.WriteHeader(http.StatusOK)
	}
}
