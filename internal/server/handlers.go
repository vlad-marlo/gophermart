package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/vlad-marlo/gophermart/pkg/luhn"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
)

const (
	BadRequestMsg  = "bad request"
	InternalErrMsg = "internal server error"
)

// handleAuthRegister ...
func (s *server) handleAuthRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u *model.User
		id := middleware.GetReqID(r.Context())

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.logger.Warnf("%v | handle auth register: close body: %v", id, err)
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
			if errors.Is(err, store.ErrLoginAlreadyInUse) {
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

// handleAuthLogin ...
func (s *server) handleAuthLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *model.User
		id := middleware.GetReqID(r.Context())

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.logger.WithField(RequestIDLoggerField, id).Warnf("%v | handle auth login: close body: %v", id, err)
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
			s.error(w, fmt.Errorf("login: unauthorized: %v", err), store.ErrIncorrectLoginData.Error(), id, http.StatusUnauthorized)
			return
		}
		if err := user.ComparePassword(req.Password); err != nil {
			s.error(w, fmt.Errorf("login: compare pass: unauthorized: %v", err), store.ErrIncorrectLoginData.Error(), id, http.StatusUnauthorized)
			return
		}

		s.logger.WithFields(logrus.Fields{
			UserIDLoggerField: user.ID,
			"request_id":      id,
		}).Trace("successful authenticated")

		s.authenticate(w, user.ID)
		w.WriteHeader(http.StatusOK)
	}
}

// handleOrdersPost ...
func (s *server) handleOrdersPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.logger.WithField("request_id", reqID).Warnf("handle orders post: close body: %v", err)
			}
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusBadRequest)
			return
		}
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		strNum := string(data)
		if strNum == "" {
			s.error(w, fmt.Errorf("bad request data"), "", reqID, http.StatusBadRequest)
			return
		}
		s.logger.WithFields(fields).Trace("successful read data")

		num, err := strconv.Atoi(strNum)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusBadRequest)
			return
		}
		s.logger.WithFields(fields).Trace("successful get number")

		if ok := luhn.Valid(num); !ok {
			s.error(w, err, "", reqID, http.StatusUnprocessableEntity)
			return
		}
		s.logger.WithFields(fields).Trace("request is valid")

		u, err := GetUserIDFromRequest(r)
		s.logger.WithFields(fields).Tracef("get user id %d %v", u, err)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusUnauthorized)
			return
		}
		if err := s.poller.Register(r.Context(), u, num); err != nil {
			s.logger.WithFields(fields).Tracef("poller register err", err)
			if errors.Is(err, store.ErrAlreadyRegisteredByUser) {
				w.WriteHeader(http.StatusOK)
				return
			} else if errors.Is(err, store.ErrAlreadyRegisteredByAnotherUser) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// handleOrdersGet ...
func (s *server) handleOrdersGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())

		u, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}

		ordrs, err := s.store.Order().GetAllByUser(r.Context(), u)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}
		if len(ordrs) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		data, err := json.Marshal(&ordrs)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

// handleBalanceGet ...
func (s *server) handleBalanceGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}

		b, err := s.store.User().GetBalance(r.Context(), id)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
		}
		data, err := json.Marshal(b)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(data); err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
		}
		s.logger.WithFields(map[string]interface{}{
			"request_id": reqID,
			"user":       id,
		})
	}
}

// handleBalanceWithdrawPost ...
func (s *server) handleBalanceWithdrawPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)
		user, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusUnauthorized)
			return
		}
		var withdraw *model.Withdraw
		defer func() {
			if err := r.Body.Close(); err != nil {
				s.error(w, err, "", reqID, http.StatusInternalServerError)
			}
		}()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(data, &withdraw); err != nil {
			s.error(w, err, "", reqID, http.StatusBadRequest)
			return
		}
		if err := s.store.Withdraws().Withdraw(ctx, user, withdraw); err != nil {
			if errors.Is(err, store.ErrPaymentRequired) {
				s.error(w, err, "", reqID, http.StatusPaymentRequired)
				return
			}
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// handleGetAllWithdraws ...
func (s *server) handleGetAllWithdraws() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}

		withdrawals, err := s.store.Withdraws().GetAllByUser(ctx, id)
		if err != nil {
			if errors.Is(err, store.ErrNoContent) {
				s.error(w, err, "", reqID, http.StatusNoContent)
				return
			}

			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(withdrawals)
		if err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			s.error(w, err, "", reqID, http.StatusInternalServerError)
		}
	}
}
