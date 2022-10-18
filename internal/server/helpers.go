package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vlad-marlo/gophermart/pkg/encryptor"
)

// error ...
func (s *server) error(w http.ResponseWriter, err error, msg, id string, status int) {
	http.Error(w, msg, status)
	s.logger.WithField(RequestIDLoggerField, id).Error(err)
}

// GetUserIDFromRequest ...
func GetUserIDFromRequest(r *http.Request) (int, error) {
	user, err := r.Cookie(UserIDCookieName)
	var to string
	if err != nil {
		return 0, fmt.Errorf("get cookie: %v", err)
	} else if err = encryptor.Decode(user.Value, &to); err != nil {
		return 0, fmt.Errorf("encryptor: decode: %v", err)
	}
	num, err := strconv.Atoi(to)
	if err != nil {
		return 0, fmt.Errorf("strconv Atoi: %v", err)
	}
	return num, nil
}
