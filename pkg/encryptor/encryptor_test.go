package encryptor_test

import (
	"github.com/google/uuid"
	"github.com/vlad-marlo/gophermart/pkg/encryptor"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"testing"
)

func TestEncryptor_DecodeEncode(t *testing.T) {
	var data []string
	defer logger.DeleteLogFolderAndFile()

	for i := 0; i < 100; i++ {
		data = append(data, uuid.New().String())
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
