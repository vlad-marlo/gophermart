package server

import (
	"github.com/google/uuid"
	"testing"
)

func TestEncryptor_DecodeEncode(t *testing.T) {
	var data []string
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
