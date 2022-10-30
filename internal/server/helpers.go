package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/pkg/encryptor"
)

// error ...
func (s *server) error(w http.ResponseWriter, err error, fields map[string]interface{}, status int) {
	w.WriteHeader(status)
	var lvl logrus.Level
	switch {
	case status >= 500:
		lvl = logrus.ErrorLevel
	case status >= 400:
		lvl = logrus.DebugLevel
	default:
		lvl = logrus.TraceLevel
	}
	s.logger.WithFields(fields).Log(lvl, err)
}

// GetUserIDFromRequest ...
func GetUserIDFromRequest(r *http.Request) (int, error) {
	user, err := r.Cookie(UserIDCookieName)
	to := new(string)
	if err != nil {
		return 0, fmt.Errorf("get cookie: %v", err)
	}
	if err := encryptor.Decode(user.Value, to); err != nil {
		return 0, fmt.Errorf("encryptor: decode: %v", err)
	}
	num, err := strconv.Atoi(*to)
	if err != nil {
		return 0, fmt.Errorf("strconv Atoi: %v", err)
	}
	return num, nil
}
