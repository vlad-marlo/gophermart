package server

import (
	"encoding/json"
	"errors"
	"fmt"
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
		id, err := GetIDFromContext(r.Context())
		if err != nil {
			s.logger.Warnf("handle auth register: get id from context: %v", err)
		}

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.logger.Tracef("%v | handle auth register: close body: %v", id, err)
			}
		}()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("%v | auth: register: read data from request: %v", id, err), InternalErrMsg, http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &u); err != nil {
			s.error(w, fmt.Errorf("%v| json unmarshal: %v", id, err), BadRequestMsg, http.StatusBadRequest)
			return
		}

		if err := u.BeforeCreate(); err != nil {
			s.error(w, fmt.Errorf("%v | user: before create: %v", id, err), InternalErrMsg, http.StatusInternalServerError)
			return
		}

		if err := s.store.User().Create(r.Context(), u); err != nil {
			if errors.Is(err, sqlstore.ErrLoginAlreadyInUse) {
				s.error(w, fmt.Errorf("%v | auth register: create user: %v", id, err), err.Error(), http.StatusConflict)
				return
			}
			s.error(w, fmt.Errorf("%v | auth register: create user: %v", id, err), InternalErrMsg, http.StatusInternalServerError)
			return
		}

		s.authenticate(w, u.ID)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) handleAuthLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *model.User

		id, err := GetIDFromContext(r.Context())
		if err != nil {
			s.logger.Warnf("handle auth login: get id from context: %v", err)
		}

		defer func() {
			if err := r.Body.Close(); err != nil {
				s.logger.Tracef("%v | handle auth login: close body: %v", id, err)
			}
		}()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("%v | auth: register: read data from request: %v", id, err), "internal server error", http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &req); err != nil {
			s.error(w, fmt.Errorf("%v | login: uncorrect request data: %v", id, err), "bad request", http.StatusBadRequest)
			return
		}

		user, err := s.store.User().GetByLogin(r.Context(), req.Login)
		if err != nil {
			s.error(w, fmt.Errorf("%v | login: unauthorized: %v", id, err), sqlstore.ErrUncorrectLoginData.Error(), http.StatusUnauthorized)
			return
		}
		if err := user.ComparePassword(req.Password); err != nil {
			s.error(w, fmt.Errorf("%v | login: compare pass: unauthorized: %v", id, err), sqlstore.ErrUncorrectLoginData.Error(), http.StatusUnauthorized)
			return
		}
		s.authenticate(w, user.ID)
		w.WriteHeader(http.StatusOK)
	}
}
