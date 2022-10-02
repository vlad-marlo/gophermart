package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/vlad-marlo/gophermart/internal/store/model"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

func (s *server) handleAuthRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u *model.User

		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("auth: register: read data from request: %v", err), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &u); err != nil {
			s.error(w, fmt.Errorf("json unmarshal: %v", err), http.StatusBadRequest)
			return
		}
		s.logger.Debugf("%s - %s - %s", u.Login, u.Password, u.EncryptedPassword)
		if err := u.BeforeCreate(); err != nil {
			s.error(w, fmt.Errorf("user: before create: %v", err), http.StatusInternalServerError)
			return
		}
		s.logger.Debugf("%s - %s - %s", u.Login, u.Password, u.EncryptedPassword)
		if err := s.store.User().Create(r.Context(), u); err != nil {
			if errors.Is(err, sqlstore.ErrLoginAlreadyInUse) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			s.error(w, err, http.StatusInternalServerError)
			return
		}
		s.authentificate(w, u.ID)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) handleAuthLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *model.User

		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("auth: register: read data from request: %v", err), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(data, &req); err != nil {
			s.error(w, fmt.Errorf("login: uncorrect request data: %v", err), http.StatusBadRequest)
			return
		}
		s.logger.Debugf("%v", req.EncryptedPassword)

		user, err := s.store.User().GetByLogin(r.Context(), req.Login)
		if err != nil {
			if errors.Is(err, sqlstore.ErrUncorrectLoginData) {
				s.error(w, fmt.Errorf("login: unauthorized: %v", err), http.StatusUnauthorized)
				return
			}
			s.error(w, err, http.StatusUnauthorized)
			return
		}
		if err := user.ComparePassword(req.Password); err != nil {
			s.error(w, fmt.Errorf("login: compare pass: unauthorized: %v", err), http.StatusUnauthorized)
			return
		}
		s.authentificate(w, user.ID)
		w.WriteHeader(http.StatusOK)
	}
}
