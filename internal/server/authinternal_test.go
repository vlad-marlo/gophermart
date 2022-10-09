package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
	"github.com/vlad-marlo/gophermart/internal/store"
	"net/http"
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

func Test_server_authenticate(t *testing.T) {
	type (
		fields struct {
			Router chi.Router
			store  store.Storage
			logger logger.Logger
			config *config.Config
		}

		args struct {
			w  http.ResponseWriter
			id int
		}
	)
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				Router: tt.fields.Router,
				store:  tt.fields.store,
				logger: tt.fields.logger,
				config: tt.fields.config,
			}
			s.authenticate(tt.args.w, tt.args.id)
		})
	}
}
