package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/vlad-marlo/gophermart/pkg/logger"
)

type encryptor struct {
	nonce []byte
	GCM   cipher.AEAD
}

var e *encryptor

func init() {
	key, err := generateRandom(aes.BlockSize)
	l := logger.GetLogger()
	if err != nil {
		l.Panicf("generate key: %v", err)
		return
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

	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		l.Panicf("initialize GCM nonce: %v", err)
		return
	}

	e = &encryptor{
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

func Encode(str string) string {
	src := []byte(str)
	dst := e.GCM.Seal(nil, e.nonce, src, nil)
	return hex.EncodeToString(dst)
}

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
