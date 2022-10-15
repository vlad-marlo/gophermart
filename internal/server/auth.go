package server

import (
	"fmt"
	"net/http"
	"strconv"

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
		var rawUserID string

		// check request id from request
		id := middleware.GetReqID(r.Context())

		if err := GetUserIDFromRequest(r, &rawUserID); err != nil {
			s.logger.WithField(RequestIDLoggerField, id).Debug(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		intID, err := strconv.Atoi(rawUserID)
		if err != nil {
			s.error(w, fmt.Errorf("parse id from cookie: %v", err), sqlstore.ErrUncorrectLoginData.Error(), id, http.StatusUnauthorized)
			return
		}

		if ok := s.store.User().ExistsWithID(r.Context(), intID); !ok {
			s.error(w, fmt.Errorf("auth middleware: exists with id: %v", err), InternalErrMsg, id, http.StatusInternalServerError)
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
