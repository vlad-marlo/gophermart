package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
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
	UserIDCookieName     = "user"
	RequestIDLoggerField = "request_id"
	UserIDLoggerField    = "user_id"
)

var encryptor *Encryptor

// init ...
func init() {
	key, err := generateRandom(aes.BlockSize)
	if err != nil {
		panic(fmt.Sprintf("generate key: %v", err))
		return
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Sprintf("initialize cipher: %v", err))
		return
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		panic(fmt.Sprintf("initialize GCM encryptor: %v", err))
		return
	}

	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		panic(fmt.Sprintf("initialize GCM nonce: %v", err))
		return
	}

	encryptor = &Encryptor{
		nonce: nonce,
		GCM:   aesGCM,
	}
}

func GetEncryptor() *Encryptor {
	return encryptor
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

		// check request id from request
		id := middleware.GetReqID(r.Context())

		if user, err := r.Cookie(UserIDCookieName); err != nil {

			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err = encryptor.Decode(user.Value, &rawUserID); err != nil {

			s.logger.WithField(RequestIDLoggerField, id).Warnf("decode: %v", err)
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

func (s *server) authenticate(w http.ResponseWriter, id int) {
	encoded := encryptor.Encode(fmt.Sprint(id))
	c := &http.Cookie{
		Name:  UserIDCookieName,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, c)
}
