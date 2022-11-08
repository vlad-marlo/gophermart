package server_test

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/pkg/logger"
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

	userLoginPath       = "/api/user/login"
	userBalancePath     = "/api/user/balance"
	userRegisterPath    = "/api/user/register"
	userOrdersPath      = "/api/user/orders"
	userWithdrawPath    = "/api/user/balance/withdraw"
	userWithdrawalsPath = "/api/user/balance/withdrawals"

	validOrderNum1 = 12345678903
	validOrderNum2 = 4532733309529845
	validOrderNum3 = 4539088167512356

	l = logger.GetLoggerByEntry(logrus.NewEntry(logrus.New()))
)

// testRequest ...
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body []byte, cookies []*http.Cookie) (*resty.Response, []byte) {
	var err error
	client := resty.New()
	r := client.R()
	if body != nil && method == http.MethodPost {
		r = r.SetBody(body)
	}
	if cookies != nil {
		r = r.SetCookies(cookies)
	}
	var resp *resty.Response
	switch method {
	case http.MethodPost:
		resp, err = r.Post(ts.URL + path)
	case http.MethodGet:
		resp, err = r.Get(ts.URL + path)
	default:
		t.Fatalf("got unexpected method: %s", method)
	}
	require.NoError(t, err)

	return resp, resp.Body()
}

func getUserCookies(t *testing.T, ts *httptest.Server, u *model.User) []*http.Cookie {
	data, err := json.Marshal(u)
	//t.Logf("%s", data)
	require.NoError(t, err, fmt.Sprintf("got unexpected err: %v", err))
	resp, _ := testRequest(t, ts, http.MethodPost, userRegisterPath, data, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	return resp.Cookies()
}
