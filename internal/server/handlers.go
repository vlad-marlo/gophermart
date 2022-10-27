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
		fields := map[string]interface{}{
			"request_id": id,
		}
		l := s.logger.WithFields(fields)
		fields["handler"] = "auth register"

		defer func() {
			if err := r.Body.Close(); err != nil {
				l.Warnf("handle auth register: close body: %v", err)
			}
		}()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("auth: register: read data from request: %v", err), InternalErrMsg, fields, http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &u); err != nil {
			s.error(w, fmt.Errorf("json unmarshal: %v", err), BadRequestMsg, fields, http.StatusBadRequest)
			return
		}

		if err := s.store.User().Create(r.Context(), u); err != nil {
			if errors.Is(err, store.ErrLoginAlreadyInUse) {
				s.error(w, fmt.Errorf("auth register: create user: %v", err), err.Error(), fields, http.StatusConflict)
				return
			}
			s.error(w, fmt.Errorf("auth register: create user: %v", err), InternalErrMsg, fields, http.StatusInternalServerError)
			return
		}

		l.Trace("successful authenticated")

		s.authenticate(w, u.ID)
		w.WriteHeader(http.StatusOK)
	}
}

// handleAuthLogin ...
func (s *server) handleAuthLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *model.User
		id := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": id,
		}
		l := s.logger.WithFields(fields)
		fields["handler"] = "auth login"

		defer func() {
			if err := r.Body.Close(); err != nil {
				l.Warnf("handle auth login: close body: %w", err)
			}
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("auth: register: read data from request: %v", err), "internal server error", fields, http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &req); err != nil {
			s.error(w, fmt.Errorf("login: uncorrect request data: %v", err), "bad request", fields, http.StatusBadRequest)
			return
		}
		l.Trace("successful read data and unmarshal")

		user, err := s.store.User().GetByLogin(r.Context(), req.Login)
		if err != nil {
			s.error(w, fmt.Errorf("login: unauthorized: %v", err), "", fields, http.StatusUnauthorized)
			return
		}
		l.Trace("got user by login")
		if err := user.ComparePassword(req.Password); err != nil {
			s.error(w, fmt.Errorf("login: compare pass: unauthorized: %v", err), "", fields, http.StatusUnauthorized)
			return
		}

		l.Trace("successful authenticated")

		s.authenticate(w, user.ID)
		w.WriteHeader(http.StatusOK)
	}
}

// handleOrdersPost ...
func (s *server) handleOrdersPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		l := s.logger.WithFields(fields)
		fields["handler"] = "order register"

		defer func() {
			if err := r.Body.Close(); err != nil {
				l.Warnf("handle orders post: close body: %v", err)
			}
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, err, "", fields, http.StatusBadRequest)
			return
		}

		strNum := string(data)
		if strNum == "" {
			s.error(w, fmt.Errorf("bad request data"), "", fields, http.StatusBadRequest)
			return
		}

		l.Trace("successful read data")

		num, err := strconv.Atoi(strNum)
		if err != nil {
			s.error(w, err, "", fields, http.StatusBadRequest)
			return
		}
		l.Trace("successful get number")

		if ok := luhn.Valid(num); !ok {
			s.error(w, err, "", fields, http.StatusUnprocessableEntity)
			return
		}
		l.Trace("request is valid")

		u, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", fields, http.StatusUnauthorized)
			return
		}
		l.Trace("successful get user from request")

		if err := s.poller.Register(r.Context(), u, num); err != nil {
			l.Tracef("poller register err", err)
			var status int

			if errors.Is(err, store.ErrAlreadyRegisteredByUser) {
				status = http.StatusOK
			} else if errors.Is(err, store.ErrAlreadyRegisteredByAnotherUser) {
				status = http.StatusConflict
			} else {
				status = http.StatusInternalServerError
			}

			s.error(w, err, "", fields, status)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// handleOrdersGet ...
func (s *server) handleOrdersGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		l := s.logger.WithFields(fields)
		fields["handler"] = "get user orders"

		u, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", fields, http.StatusInternalServerError)
			return
		}
		l.Trace("successful got user from request")

		orders, err := s.store.Order().GetAllByUser(r.Context(), u)
		if err != nil {
			if errors.Is(err, store.ErrNoContent) {
				s.error(w, err, "", fields, http.StatusNoContent)
				return
			}
			s.error(w, err, "", fields, http.StatusInternalServerError)
			return
		}
		l.Trace("successful get user orders")

		if len(orders) == 0 {
			s.error(w, err, "", fields, http.StatusNoContent)
			return
		}

		data, err := json.Marshal(&orders)
		if err != nil {
			s.error(w, fmt.Errorf("json marshal: %v", err), "", fields, http.StatusInternalServerError)
			return
		}
		l.Trace("successful marshaled orders to resp data")

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(data); err != nil {
			s.error(w, fmt.Errorf("write response: %v", err), "", fields, http.StatusInternalServerError)
			return
		}
	}
}

// handleBalanceGet ...
func (s *server) handleBalanceGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		l := s.logger.WithFields(fields)
		fields["handler"] = "get user balance"

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", fields, http.StatusInternalServerError)
			return
		}

		b, err := s.store.User().GetBalance(r.Context(), id)
		if err != nil {
			s.error(w, err, "", fields, http.StatusInternalServerError)
			return
		}
		l.Trace("successful got user balance", b.Current)
		data, err := json.Marshal(b)
		if err != nil {
			s.error(w, fmt.Errorf("json marshal: %v", err), "", fields, http.StatusInternalServerError)
			return
		}
		l.Trace("successful marshaled")

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(data); err != nil {
			s.error(w, fmt.Errorf("rw write: %v", err), "", fields, http.StatusInternalServerError)
		}
	}
}

// handleBalanceWithdrawPost ...
func (s *server) handleBalanceWithdrawPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		l := s.logger.WithFields(fields)
		fields["handler"] = "handle balance withdraw post"

		user, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, fmt.Errorf("get user id from req: %v", err), "", fields, http.StatusUnauthorized)
			return
		}
		l.Trace("successful get user from request")

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.error(w, fmt.Errorf("resp body close: %v", err), "", fields, http.StatusInternalServerError)
			}
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("read body data: %v", err), "", fields, http.StatusInternalServerError)
			return
		}
		l.Trace("successful get body data")

		var withdraw *model.Withdraw
		if err := json.Unmarshal(data, &withdraw); err != nil {
			s.error(w, fmt.Errorf("json unmarshal withdraw: %v", err), "", fields, http.StatusBadRequest)
			return
		}
		l.Trace("successful json unmarshalled data")

		if err := s.store.Withdraws().Withdraw(ctx, user, withdraw); err != nil {
			err = fmt.Errorf("withdraw: %v", err)
			var status int
			if errors.Is(err, store.ErrPaymentRequired) {
				status = http.StatusPaymentRequired
			} else {
				status = http.StatusInternalServerError
			}
			s.error(w, err, "", fields, status)
			return
		}
		l.Trace("successful withdraw")
		w.WriteHeader(http.StatusOK)
	}
}

// handleGetAllWithdraws ...
func (s *server) handleGetAllWithdraws() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		l := s.logger.WithFields(fields)
		fields["handler"] = "get all withdraws"

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, "", fields, http.StatusInternalServerError)
			return
		}
		l.Trace("successful get user id from req")

		withdrawals, err := s.store.Withdraws().GetAllByUser(ctx, id)
		if err != nil {
			err = fmt.Errorf("withdraws: get all by user: %v", err)
			if errors.Is(err, store.ErrNoContent) {
				s.error(w, err, "", fields, http.StatusNoContent)
				return
			}

			s.error(w, err, "", fields, http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(withdrawals)
		if err != nil {
			s.error(w, err, "", fields, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			s.error(w, err, "", fields, http.StatusInternalServerError)
		}
	}
}
