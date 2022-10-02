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
		if err := u.BeforeCreate(); err != nil {
			s.error(w, fmt.Errorf("user: before create: %v", err), http.StatusInternalServerError)
			return
		}
		if err := s.store.User().Create(r.Context(), u); err != nil {
			if errors.Is(err, sqlstore.ErrLoginAlreadyInUse) {
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		s.authentificate(w, u.ID)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) handleAuthLogin() http.HandlerFunc {
	type request struct {
		Login string `json:"login"`
		Pass  string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req *request

		data, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, fmt.Errorf("auth: register: read data from request: %v", err), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(data, &req); err != nil {
			s.error(w, fmt.Errorf("uncorrect request data: %v", err), http.StatusBadRequest)
			return
		}
		id, err := s.store.User().GetIDByLoginAndPass(r.Context(), req.Login, req.Pass)
		if err != nil {
			if errors.Is(err, sqlstore.ErrUncorrectLoginData) {
				s.error(w, fmt.Errorf(""), http.StatusUnauthorized)
				return
			}
		}
		s.authentificate(w, id)
		w.WriteHeader(http.StatusOK)
	}
}
