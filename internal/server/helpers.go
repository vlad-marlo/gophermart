package server

import (
	"fmt"
	"net/http"

	"github.com/vlad-marlo/gophermart/pkg/encryptor"
)

// error ...
func (s *server) error(w http.ResponseWriter, err error, msg, id string, status int) {
	http.Error(w, msg, status)
	s.logger.WithField(RequestIDLoggerField, id).Error(err)
}

// GetUserIDFromRequest ...
func GetUserIDFromRequest(r *http.Request, to *string) error {
	user, err := r.Cookie(UserIDCookieName)
	if err != nil {
		return fmt.Errorf("get cookie: %v", err)
	} else if err = encryptor.Decode(user.Value, to); err != nil {
		return fmt.Errorf("encryptor: decode: %v", err)
	}
	return nil
}
