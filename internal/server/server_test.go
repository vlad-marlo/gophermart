package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	conStr               = os.Getenv("TEST_DB_URI")
	userLogin1           = "server_first"
	userLogin2           = "server_second"
	userPassword         = "password"
	userTableName        = "users"
	ordersTableName      = "orders"
	withdrawalsTableName = "withdrawals"

	userLoginPath    = "/api/user/login"
	userBalancePath  = "/api/user/balance"
	userRegisterPath = "/api/user/register"
	userOrdersPath   = "/api/user/orders"
	userWithdrawPath = "/api/user/balance/withdraw"
	//userWithdrawalsPath = "/api/user/balance/withdrawals"

	validOrderNum1 = 12345678904
	validOrderNum2 = 4532733309529845
	validOrderNum3 = 4539088167512356

	l = logger.GetLoggerByEntry(logrus.NewEntry(logrus.New()))
)

// testRequest ...
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, cookies []*http.Cookie) (*http.Response, []byte) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)
	if cookies != nil {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoErrorf(t, err, "close request body: %v", err)

	return resp, respBody
}

func getUserCookies(t *testing.T, ts *httptest.Server, u *model.User) []*http.Cookie {
	data, err := json.Marshal(u)
	//t.Logf("%s", data)
	require.NoError(t, err, fmt.Sprintf("got unexpected err: %v", err))
	resp, _ := testRequest(t, ts, http.MethodPost, userRegisterPath, bytes.NewReader(data), nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer require.NoError(t, resp.Body.Close())
	return resp.Cookies()
}
