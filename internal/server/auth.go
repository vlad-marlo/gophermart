package server

import (
	"fmt"
	"net/http"

	"github.com/vlad-marlo/gophermart/pkg/encryptor"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	UserIDCookieName     = "user"
	RequestIDLoggerField = "request_id"
)

// CheckAuthMiddleware ...
func (s *server) CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check request reqID from request
		reqID := middleware.GetReqID(r.Context())
		fields := map[string]interface{}{
			"request_id": reqID,
			"middleware": "check auth middleware",
		}
		l := s.logger.WithFields(fields)

		id, err := GetUserIDFromRequest(r)
		if err != nil {
			l.Tracef("get user id from request: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err != nil {
			s.error(w, fmt.Errorf("parse id from cookie: %v", err), "", fields, http.StatusUnauthorized)
			return
		}

		if ok := s.store.User().ExistsWithID(r.Context(), id); !ok {
			s.error(w, fmt.Errorf("auth middleware: exists with id: %v", err), InternalErrMsg, fields, http.StatusInternalServerError)
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
