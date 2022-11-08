package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type encryptor struct {
	nonce []byte
	GCM   cipher.AEAD
}

var e encryptor

const (
	EnvKey      = "ENCRYPTOR_KEY"
	EnvNonceKey = "ENCRYPTOR_NONCE"
)

func init() {
	var err error
	l := logger.GetLogger()

	key, err := hex.DecodeString(os.Getenv(EnvKey))
	if err != nil {
		l.Errorf("hex decode string: %v", err)
	}

	if len(key) != aes.BlockSize {
		key, err = generateRandom(aes.BlockSize)
		if err != nil {
			l.Panicf("generate key: %v", err)
			return
		}

		_ = os.Unsetenv(EnvKey)

		if err := os.Setenv(EnvKey, hex.EncodeToString(key)); err != nil {
			l.Errorf("set env: %v", err)
		}
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		l.Panicf("initialize cipher: %v", err)
		return
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		l.Panicf("initialize GCM encryptor: %v", err)
		return
	}

	nonce, err := hex.DecodeString(os.Getenv(EnvNonceKey))
	if err != nil || len(nonce) != aesGCM.NonceSize() {
		if err != nil {
			l.Errorf("hex: decode string: nonce key: %v", err)
		}

		nonce, err = generateRandom(aesGCM.NonceSize())
		if err != nil {
			l.Panicf("initialize GCM nonce: %v", err)
			return
		}

		if err = os.Unsetenv(EnvNonceKey); err != nil {
			l.Errorf("os: unset env: %v", err)
		}

		if err := os.Setenv(EnvNonceKey, hex.EncodeToString(nonce)); err != nil {
			l.Errorf("set env nonce: %v", err)
		}
	}

	e = encryptor{
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
func Encode(str string) string {
	src := []byte(str)
	dst := e.GCM.Seal(nil, e.nonce, src, nil)
	return hex.EncodeToString(dst)
}

// Decode ...
func Decode(str string, to *string) error {
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
