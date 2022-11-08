package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/luhn"
	"io"
	"net/http"
	"strconv"
)

// handleAuthRegister ...
func (s *Server) handleAuthRegister() http.HandlerFunc {
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
			s.error(w, fmt.Errorf("auth: register: read data from request: %w", err), fields, http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &u); err != nil {
			s.error(w, fmt.Errorf("json unmarshal: %w", err), fields, http.StatusBadRequest)
			return
		}
		if !u.Valid() {
			s.error(w, err, fields, http.StatusBadRequest)
			return
		}

		if err := s.store.User().Create(r.Context(), u); err != nil {
			if errors.Is(err, store.ErrLoginAlreadyInUse) {
				s.error(w, fmt.Errorf("auth register: create user: %w", err), fields, http.StatusConflict)
				return
			}
			s.error(w, fmt.Errorf("auth register: create user: %w", err), fields, http.StatusInternalServerError)
			return
		}

		s.authenticate(w, u.ID)
		w.WriteHeader(http.StatusOK)
	}
}

// handleAuthLogin ...
func (s *Server) handleAuthLogin() http.HandlerFunc {
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
			s.error(w, fmt.Errorf("auth: register: read data from request: %w", err), fields, http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &req); err != nil {
			s.error(w, fmt.Errorf("login: uncorrect request data: %w", err), fields, http.StatusBadRequest)
			return
		}

		if req.Login == "" || req.Password == "" {
			s.error(w, errors.New("bad request"), fields, http.StatusBadRequest)
			return
		}

		user, err := s.store.User().GetByLogin(r.Context(), req.Login)
		if err != nil {
			if errors.Is(err, store.ErrIncorrectLoginData) {
				s.error(w, fmt.Errorf("login: unauthorized: %w", err), fields, http.StatusUnauthorized)
				return
			}
			s.error(w, fmt.Errorf("get user by login: %w", err), fields, http.StatusInternalServerError)

			return
		}
		if err := user.ComparePassword(req.Password); err != nil {
			s.error(w, fmt.Errorf("login: compare pass: unauthorized: %w", err), fields, http.StatusUnauthorized)
			return
		}

		s.authenticate(w, user.ID)
		w.WriteHeader(http.StatusOK)
	}
}

// handleOrdersPost ...
func (s *Server) handleOrdersPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// logging stuff
		reqID := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		l := s.logger.WithFields(fields)

		u, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, fields, http.StatusUnauthorized)
			return
		}

		var data []byte
		defer func() {
			if err := r.Body.Close(); err != nil {
				l.Warn(fmt.Sprintf("close body: %v", err))
			}
		}()
		data, err = io.ReadAll(r.Body)
		if err != nil {
			s.error(w, err, fields, http.StatusInternalServerError)
			return
		}

		var num int
		num, err = strconv.Atoi(string(data))
		if err != nil {
			s.error(w, err, fields, http.StatusBadRequest)
			return
		}

		if !luhn.Valid(num) {
			s.error(w, fmt.Errorf("bad number: was not pass luhn test"), fields, http.StatusUnprocessableEntity)
			return
		}

		if err := s.store.Order().Register(r.Context(), u, num); err != nil {
			switch {
			case errors.Is(err, store.ErrAlreadyRegisteredByAnotherUser):
				s.error(w, err, fields, http.StatusConflict)
			case errors.Is(err, store.ErrAlreadyRegisteredByUser):
				s.error(w, err, fields, http.StatusOK)
			default:
				s.error(w, err, fields, http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// handleOrdersGet ...
func (s *Server) handleOrdersGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		fields["handler"] = "get user orders"

		var (
			orders []*model.Order
			data   []byte
		)

		w.Header().Set("Content-Type", "application/json")

		u, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, fields, http.StatusUnauthorized)
			return
		}

		orders, err = s.store.Order().GetAllByUser(r.Context(), u)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNoContent):
				s.error(w, err, fields, http.StatusNoContent)
			default:
				s.error(w, err, fields, http.StatusInternalServerError)
			}
			return
		}

		data, err = json.Marshal(orders)
		if err != nil {
			s.error(w, fmt.Errorf("json marshal: %w", err), fields, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			err = fmt.Errorf("write response: %w", err)
			s.error(w, err, fields, http.StatusInternalServerError)
		}
	}
}

// handleBalanceGet ...
func (s *Server) handleBalanceGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		reqID := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": reqID,
		}
		fields["handler"] = "get user balance"

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, fields, http.StatusInternalServerError)
			return
		}

		b, err := s.store.User().GetBalance(r.Context(), id)
		if err != nil {
			s.error(w, err, fields, http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(b)
		if err != nil {
			s.error(w, fmt.Errorf("json marshal: %v", err), fields, http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(data); err != nil {
			s.error(w, fmt.Errorf("response writer: write data: %v", err), fields, http.StatusInternalServerError)
		}
	}
}

// handleGetAllWithdraws ...
func (s *Server) handleGetAllWithdraws() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)
		fields := map[string]interface{}{
			"request_id": reqID,
			"handler":    "get all withdraws",
		}

		w.Header().Set("Content-Type", "application/json")

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, fields, http.StatusInternalServerError)
			return
		}

		withdrawals, err := s.store.Withdraws().GetAllByUser(ctx, id)
		if err != nil {
			err = fmt.Errorf("withdraws: get all by user: %w", err)

			if errors.Is(err, store.ErrNoContent) {
				s.error(w, err, fields, http.StatusNoContent)
				return
			}

			s.error(w, err, fields, http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(withdrawals)
		if err != nil {
			s.error(w, err, fields, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			s.error(w, err, fields, http.StatusInternalServerError)
		}
	}
}

// handleWithdrawsPost ...
func (s *Server) handleWithdrawsPost() http.HandlerFunc {
	type request struct {
		Order int     `json:"order,string"`
		Sum   float64 `json:"sum"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		fields := map[string]interface{}{
			"request_id": middleware.GetReqID(ctx),
		}
		l := s.logger.WithFields(fields)

		u, err := GetUserIDFromRequest(r)
		if err != nil {
			s.error(w, err, fields, http.StatusUnauthorized)
			return
		}

		defer func() {
			if err := r.Body.Close(); err != nil {
				l.Errorf(fmt.Sprintf("request body close: %v", err))
			}
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, err, fields, http.StatusInternalServerError)
			return
		}

		var req *request
		if err := json.Unmarshal(data, &req); err != nil {
			s.error(w, err, fields, http.StatusBadRequest)
			return
		}

		withdraw := &model.Withdraw{
			Order: req.Order,
			Sum:   req.Sum,
		}

		if err := s.store.Withdraws().Withdraw(ctx, u, withdraw); err != nil {
			err = fmt.Errorf("withdraw: %w", err)
			switch {
			case errors.Is(err, store.ErrIncorrectData):
				s.error(w, err, fields, http.StatusUnprocessableEntity)
			case errors.Is(err, store.ErrPaymentRequired):
				s.error(w, err, fields, http.StatusPaymentRequired)
			default:
				s.error(w, err, fields, http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
