package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
)

type (
	Encryptor struct {
		nonce []byte
		GCM   cipher.AEAD
	}
	cookieUserIDValueType string
)

const (
	UserIDCookieName = "user"
)

var encryptor *Encryptor

// generateRandom byte slice
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("rand read: %v", err)
	}
	return b, nil
}

// NewEncryptor ...
func NewEncryptor() error {
	if encryptor != nil {
		return nil
	}

	key, err := generateRandom(aes.BlockSize)
	if err != nil {
		return fmt.Errorf("generate key: %v", err)
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("initialize cipher: %v", err)
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return fmt.Errorf("initialize GCM encryptor: %v", err)
	}

	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		return fmt.Errorf("initialize GCM nonce: %v", err)
	}

	encryptor = &Encryptor{
		nonce: nonce,
		GCM:   aesGCM,
	}

	return nil
}

// EncodeUUID ...
func (e *Encryptor) EncodeUUID(uuid string) string {
	src := []byte(uuid)
	dst := e.GCM.Seal(nil, e.nonce, src, nil)
	return hex.EncodeToString(dst)
}

// DecodeUUID ...
func (e *Encryptor) DecodeUUID(uuid string, to *string) error {
	dst, err := hex.DecodeString(uuid)
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

		if err := NewEncryptor(); err != nil {
			s.logger.Warnf("new encryptor: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user, err := r.Cookie(UserIDCookieName); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err = encryptor.DecodeUUID(user.Value, &rawUserID); err != nil {
			s.logger.Warnf("decode: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		intID, err := strconv.Atoi(rawUserID)
		if err != nil {
			s.error(w, fmt.Errorf("parse id from cookie: %v", err), http.StatusUnauthorized)
			return
		}
		if ok, err := s.store.User().ExistsWithID(r.Context(), intID); err != nil {
			s.error(w, fmt.Errorf("auth middleware: exists with id: %v", err), http.StatusInternalServerError)
			return
		} else if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) authentificate(w http.ResponseWriter, id int) {
	if err := NewEncryptor(); err != nil {
		s.logger.Warnf("auth: new encryptor: %s", err)
	}
	encoded := encryptor.EncodeUUID(fmt.Sprint(id))
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
	if err := NewEncryptor(); err != nil {
		return "", fmt.Errorf("new encryptor: %v", err)
	}
	if err := encryptor.DecodeUUID(user.Value, &userID); err != nil {
		return "", fmt.Errorf("decode: %v", err)
	}
	return
}
