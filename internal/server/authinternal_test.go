package server_test

import (
	"github.com/google/uuid"
	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
	"github.com/vlad-marlo/gophermart/internal/server"
	"testing"
)

func TestEncryptor_DecodeEncode(t *testing.T) {
	var data []string
	encryptor := server.GetEncryptor()
	defer logger.DeleteLogFolderAndFile()

	for i := 0; i < 100; i++ {
		data = append(data, uuid.New().String())
	}

	if encryptor == nil {
		t.Fatal("encryptor must be not nil obj")
	}

	for _, v := range data {
		var decodeTo string
		encrypted := encryptor.Encode(v)
		if err := encryptor.Decode(encrypted, &decodeTo); err != nil {
			t.Errorf("encryptor: decode: %v", err)
		}
		if decodeTo != v {
			t.Errorf("decodeTo != v; %s != %s", decodeTo, v)
		}
	}
}
