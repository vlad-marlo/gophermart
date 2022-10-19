package server

import (
	"fmt"
	"net/http"

	"github.com/vlad-marlo/gophermart/pkg/encryptor"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

const (
	UserIDCookieName     = "user"
	RequestIDLoggerField = "request_id"
	UserIDLoggerField    = "user_id"
)

// CheckAuthMiddleware ...
func (s *server) CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check request reqID from request
		reqID := middleware.GetReqID(r.Context())

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			s.logger.WithField(RequestIDLoggerField, reqID).Debug(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err != nil {
			s.error(w, fmt.Errorf("parse id from cookie: %v", err), sqlstore.ErrUncorrectLoginData.Error(), reqID, http.StatusUnauthorized)
			return
		}

		if ok := s.store.User().ExistsWithID(r.Context(), id); !ok {
			s.error(w, fmt.Errorf("auth middleware: exists with id: %v", err), InternalErrMsg, reqID, http.StatusInternalServerError)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authenticate ...
func (s *server) authenticate(w http.ResponseWriter, id int) {
	encoded := encryptor.Encode(fmt.Sprint(id))
	c := &http.Cookie{
		Name:  UserIDCookieName,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, c)
}
