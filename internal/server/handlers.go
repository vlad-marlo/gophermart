package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vlad-marlo/gophermart/internal/pkg/luhn"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"

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

func (s *server) handleOrdersPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())

		defer func() {
			_ = r.Body.Close()
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusBadRequest)
			return
		}
		strNum := string(data)
		if strNum == "" {
			s.error(w, fmt.Errorf("bad request data"), "", reqID, http.StatusBadRequest)
			return
		}

		num, err := strconv.Atoi(strNum)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusBadRequest)
			return
		}

		if ok := luhn.Valid(num); !ok {
			s.error(w, err, "", reqID, http.StatusUnprocessableEntity)
			return
		}
		_, _ = w.Write([]byte(strNum))
	}
}

func (s *server) handleOrdersGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO implement me
		_, _ = w.Write([]byte("orders get"))
	}
}

func (s *server) handleBalanceGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO implement me
		_, _ = w.Write([]byte("balance get"))
	}
}

func (s *server) handleBalanceWithdrawPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO implement me
		_, _ = w.Write([]byte("withdraw balance"))
	}
}

func (s *server) handleGetAllWithdraws() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO implement me
		_, _ = w.Write([]byte("get all withdraws"))
	}
}
