package middlewares

import (
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRecoverer ...
func TestRecoverer(t *testing.T) {
	defer logger.DeleteLogFolderAndFile()
	r := TestServer(t)
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
	require.NoErrorf(t, err, "new request: %v", err)

	resp, err := http.DefaultClient.Do(req)
	require.NoErrorf(t, err, "do request: %v", err)

	require.Equal(t, 500, resp.StatusCode)
	require.Nil(t, recover(), "recoverer is not nil!?!?!?!?")
}
