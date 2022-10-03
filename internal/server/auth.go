package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
	"net/http"
	"strconv"
)

type (
	Encryptor struct {
		nonce []byte
		GCM   cipher.AEAD
	}
)

const (
	UserIDCookieName = "user"
)

var encryptor *Encryptor

// init ...
func init() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	if encryptor != nil {
		return
	}

	key, err := generateRandom(aes.BlockSize)
	if err != nil {
		logger.Fatalf("generate key: %v", err)
		return
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		logger.Fatalf("initialize cipher: %v", err)
		return
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		logger.Fatalf("initialize GCM encryptor: %v", err)
		return
	}

	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		logger.Fatalf("initialize GCM nonce: %v", err)
		return
	}

	encryptor = &Encryptor{
		nonce: nonce,
		GCM:   aesGCM,
	}
}

// generateRandom byte slice with size
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("rand read: %v", err)
	}
	return b, nil
}

// Encode ...
func (e *Encryptor) Encode(str string) string {
	src := []byte(str)
	dst := e.GCM.Seal(nil, e.nonce, src, nil)
	return hex.EncodeToString(dst)
}

// Decode ...
func (e *Encryptor) Decode(str string, to *string) error {
	dst, err := hex.DecodeString(str)
	if err != nil {
		return fmt.Errorf("hex decode: %v", err)
	}

	src, err := e.GCM.Open(nil, e.nonce, dst, nil)
	if err != nil {
		return fmt.Errorf("gcm open: %v", err)
	}

	*to = string(src)
	return nil
}

// CheckAuthMiddleware ...
func (s *server) CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rawUserID string

		if user, err := r.Cookie(UserIDCookieName); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err = encryptor.Decode(user.Value, &rawUserID); err != nil {
			s.logger.Warnf("decode: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		intID, err := strconv.Atoi(rawUserID)
		if err != nil {
			s.error(w, fmt.Errorf("parse id from cookie: %v", err), sqlstore.ErrUncorrectLoginData.Error(), http.StatusUnauthorized)
			return
		}
		if ok, err := s.store.User().ExistsWithID(r.Context(), intID); err != nil {
			s.error(w, fmt.Errorf("auth middleware: exists with id: %v", err), InternalErrMsg, http.StatusInternalServerError)
			return
		} else if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) authenticate(w http.ResponseWriter, id int) {
	encoded := encryptor.Encode(fmt.Sprint(id))
	c := &http.Cookie{
		Name:  UserIDCookieName,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, c)
}

func (s *server) getUserIDFromRequest(r *http.Request) (userID string, err error) {
	user, err := r.Cookie(UserIDCookieName)
	if err != nil {
		return "", fmt.Errorf("get cookie from req: %v", err)
	}
	if err := encryptor.Decode(user.Value, &userID); err != nil {
		return "", fmt.Errorf("decode: %v", err)
	}
	return
}
