package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/model"
	"github.com/vlad-marlo/gophermart/internal/poller"
	"github.com/vlad-marlo/gophermart/internal/server"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"github.com/vlad-marlo/gophermart/pkg/luhn"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthUserRegister_MainCases(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}

	type (
		request struct {
			Login    string `json:"login,omitempty"`
			Password string `json:"password,omitempty"`
		}
		onlyLoginRequest struct {
			Login string `json:"login"`
		}
		onlyPasswordRequest struct {
			Password string `json:"password"`
		}
	)

	storage, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName, ordersTableName, withdrawalsTableName)

	cfg := config.TestConfig(t)

	s := server.New(l, storage, cfg, nil)

	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	tests := []struct {
		name    string
		request func() ([]byte, error)
		code    int
	}{
		{
			name: "positive case #1",
			request: func() ([]byte, error) {
				r := &request{
					Login:    userLogin1,
					Password: userPassword,
				}
				return json.Marshal(r)
			},
			code: http.StatusOK,
		},
		{
			name: "positive case #2",
			request: func() ([]byte, error) {
				r := &request{
					Login:    userLogin2,
					Password: userPassword,
				}
				return json.Marshal(r)
			},
			code: http.StatusOK,
		},
		{
			name: "negative case #1: conflict",
			request: func() ([]byte, error) {
				r := &request{
					Login:    userLogin1,
					Password: userPassword,
				}
				return json.Marshal(r)
			},
			code: http.StatusConflict,
		},
		{
			name: "negative case #2: conflict",
			request: func() ([]byte, error) {
				r := &request{
					Login:    userLogin2,
					Password: userPassword,
				}
				return json.Marshal(r)
			},
			code: http.StatusConflict,
		},
		{
			name: "negative case #3: bad request",
			request: func() ([]byte, error) {
				r := new(onlyPasswordRequest)
				r.Password = userPassword
				return json.Marshal(r)
			},
			code: http.StatusBadRequest,
		},
		{
			name: "negative case #4: bad request",
			request: func() ([]byte, error) {
				r := new(onlyLoginRequest)
				r.Login = userLogin2
				return json.Marshal(r)
			},
			code: http.StatusBadRequest,
		},
		{
			name: "negative case #5: bad request",
			request: func() ([]byte, error) {
				return nil, nil
			},
			code: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := tt.request()
			require.NoErrorf(t, err, "json marshal: %w", err)

			resp, _ := testRequest(t, ts, http.MethodPost, userRegisterPath, bytes.NewReader(data), nil)
			assert.Equalf(t, tt.code, resp.StatusCode, "status codes are not equal: got %d: want %d", resp.StatusCode, tt.code)
		})
	}
}

func TestAuthUserRegister_CheckAuth(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}
	type request struct {
		Login    string `json:"login,omitempty"`
		Password string `json:"password,omitempty"`
	}

	tests := []struct {
		name    string
		request func() *request
	}{
		{
			name: "positive case #1",
			request: func() *request {
				return &request{
					Login:    userLogin1,
					Password: userPassword,
				}
			},
		},
	}

	storage, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName)

	cfg := config.TestConfig(t)
	s := server.New(l, storage, cfg, nil)
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request())
			require.NoErrorf(t, err, "json marshal: %w", err)
			resp, _ := testRequest(t, ts, http.MethodPost, userRegisterPath, bytes.NewReader(data), nil)
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "status codes are not equal: got %d: want %d", resp.StatusCode, 200)
			cookies := resp.Cookies()

			resp, _ = testRequest(t, ts, http.MethodGet, userBalancePath, nil, cookies)
			require.NotEqual(t, resp.StatusCode, http.StatusUnauthorized, "status code is 401")

			// check that cookies was necessary to auth
			resp, _ = testRequest(t, ts, http.MethodGet, userBalancePath, nil, nil)
			require.Equal(t, resp.StatusCode, http.StatusUnauthorized, "status code is 401")
		})
	}
}

func TestAuthUserLogin(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}
	ctx := context.Background()

	type (
		request struct {
			Login    string `json:"login,omitempty"`
			Password string `json:"password,omitempty"`
		}
		onlyLoginRequest struct {
			Login string `json:"login"`
		}
		onlyPasswordRequest struct {
			Password string `json:"password"`
		}
	)
	u := &model.User{
		Login:    userLogin1,
		Password: userPassword,
	}

	tests := []struct {
		name    string
		request func() ([]byte, error)
		// what code must be after test case
		code int
	}{
		{
			name: "positive case #1",
			request: func() ([]byte, error) {
				r := &request{
					Login:    userLogin1,
					Password: userPassword,
				}
				return json.Marshal(r)
			},
			code: http.StatusOK,
		},
		{
			name: "negative case #1: bad username",
			request: func() ([]byte, error) {
				r := &request{
					Login:    userLogin2,
					Password: userPassword,
				}
				return json.Marshal(r)
			},
			code: http.StatusUnauthorized,
		},
		{
			name: "negative case #2: bad request no login",
			request: func() ([]byte, error) {
				r := new(onlyPasswordRequest)
				r.Password = userPassword
				return json.Marshal(r)
			},
			code: http.StatusBadRequest,
		},
		{
			name: "negative case #3: bad request no password",
			request: func() ([]byte, error) {
				r := new(onlyLoginRequest)
				r.Login = userLogin2
				return json.Marshal(r)
			},
			code: http.StatusBadRequest,
		},
		{
			name: "negative case #5: bad request no auth data",
			request: func() ([]byte, error) {
				return nil, nil
			},
			code: http.StatusBadRequest,
		},
	}

	storage, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName)

	err := storage.User().Create(ctx, u)
	require.NoErrorf(t, err, "create user: %v", err)

	cfg := config.TestConfig(t)
	s := server.New(l, storage, cfg, nil)
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.request()
			require.NoErrorf(t, err, "json marshal: %w", err)
			var reader io.Reader
			if data != nil {
				reader = bytes.NewReader(data)
			} else {
				reader = nil
			}
			// doing first request
			resp, _ := testRequest(t, ts, http.MethodPost, userLoginPath, reader, nil)
			require.Equalf(t, tt.code, resp.StatusCode, "want=%d actual=%d", tt.code, resp.StatusCode)

			cookies := resp.Cookies()
			resp, _ = testRequest(t, ts, http.MethodGet, userBalancePath, nil, cookies)
			if tt.code != http.StatusOK {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "status code is 401")
			} else {
				require.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "status code is not 401 got %d", resp.StatusCode)
			}
		})
	}
}

func TestOrdersPost(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}

	cfg := config.TestConfig(t)

	storage, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName, ordersTableName)

	log := logrus.New()
	log.Out = io.Discard

	p := poller.TestPoller(t, storage)

	s := server.New(logger.GetLoggerByEntry(logrus.NewEntry(log)), storage, cfg, p)
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	u1 := &model.User{Login: userLogin1, Password: userPassword}
	u2 := &model.User{Login: userLogin2, Password: userPassword}

	cookiesUser1 := getUserCookies(t, ts, u1)
	cookiesUser2 := getUserCookies(t, ts, u2)

	tests := []struct {
		name    string
		request int
		code    int
		cookies []*http.Cookie
	}{
		{
			name:    "positive case #1",
			request: 12345678903,
			code:    http.StatusAccepted,
			cookies: cookiesUser1,
		},
		{
			name:    "negative conflict #1",
			request: 12345678903,
			code:    http.StatusConflict,
			cookies: cookiesUser2,
		},
		{
			name:    "positive case #2",
			request: 12345678903,
			code:    http.StatusOK,
			cookies: cookiesUser1,
		},
		{
			name:    "positive case #3",
			request: 1234562,
			code:    0,
			cookies: cookiesUser1,
		},
		{
			name:    "positive case #4",
			request: 12345123,
			code:    0,
			cookies: cookiesUser1,
		},
		{
			name:    "positive case #5",
			request: 12345675,
			code:    0,
			cookies: cookiesUser1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, http.MethodPost, userOrdersPath, strings.NewReader(fmt.Sprint(tt.request)), tt.cookies)
			if tt.code != 0 {
				require.Equal(t, tt.code, resp.StatusCode, fmt.Sprintf("got unexpected status code want=%d got=%d", tt.code, resp.StatusCode))
			} else if luhn.Valid(tt.request) {
				require.NotEqual(t, http.StatusUnprocessableEntity, resp.StatusCode, fmt.Sprintf("got unexpected status code: %d", resp.StatusCode))
			}
		})
	}
}

func TestOrdersGet(t *testing.T) {
	if conStr == "" {
		t.Skip("connect string is not provided")
	}

	cfg := config.TestConfig(t)
	ctx := context.Background()

	storage, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName, ordersTableName)

	log := logrus.New()
	log.Out = io.Discard

	p := poller.TestPoller(t, storage)

	s := server.New(logger.GetLoggerByEntry(logrus.NewEntry(log)), storage, cfg, p)
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	var err error

	u1 := &model.User{Login: userLogin1, Password: userPassword}
	//t.Logf("user id=%d password=%s enc_pass=%s", u1.ID, u1.Password, u1.EncryptedPassword)
	require.NoError(t, err, fmt.Sprintf("got unexpected error while creating user: %v", err))
	u1.Password = userPassword
	cookiesU1 := getUserCookies(t, ts, u1)
	u1, err = storage.User().GetByLogin(ctx, u1.Login)
	require.NoError(t, err, fmt.Sprintf("get user by login: %v", err))

	u2 := &model.User{Login: userLogin2, Password: userPassword}
	require.NoError(t, err, fmt.Sprintf("got unexpected error while creating user: %v", err))
	u2.Password = userPassword
	cookiesU2 := getUserCookies(t, ts, u2)
	u2, err = storage.User().GetByLogin(ctx, u2.Login)
	require.NoError(t, err, fmt.Sprintf("get user by login: %v", err))

	tests := []struct {
		name   string
		o      int
		u      *model.User
		otherU *model.User
	}{
		{
			name:   "positive case #1",
			o:      12345678903,
			u:      u1,
			otherU: u2,
		},
		{
			name:   "positive case #2",
			o:      79927398713,
			u:      u1,
			otherU: u2,
		},
		{
			name:   "positive case #2",
			o:      4532733309529845,
			u:      u1,
			otherU: u2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cookies, otherCookies []*http.Cookie
			if u1 == tt.u {
				cookies = cookiesU1
				otherCookies = cookiesU2
			} else {
				cookies = cookiesU2
				otherCookies = cookiesU1
			}
			resp, _ := testRequest(t, ts, http.MethodPost, userOrdersPath, strings.NewReader(fmt.Sprint(tt.o)), cookies)
			require.True(t, http.StatusAccepted == resp.StatusCode || resp.StatusCode == http.StatusOK, fmt.Sprintf("got unexpected status code want=200/202 got=%d", resp.StatusCode))
			// check exist in storage
			{
				// check order exists in user orders
				{
					orders, err := storage.Order().GetAllByUser(ctx, tt.u.ID)
					require.NoError(t, err, "get all orders by user: %v", err)

					status := false
					for _, o := range orders {
						if o.Number == tt.o {
							status = true
						}
					}
					require.True(t, status, "order was not in registered orders by user")
				}

				// check order doesn't exist in different user's orders
				{
					orders, err := storage.Order().GetAllByUser(ctx, tt.otherU.ID)
					if !errors.Is(err, store.ErrNoContent) {
						require.NoError(t, err, fmt.Sprintf("got unexpected error: %v", err))
					}

					status := false
					for _, o := range orders {
						if o.Number == tt.o {
							status = true
						}
					}
					assert.False(t, status, "found order in registered orders by another user")
				}
			}
			// check exist in get http
			{
				// check contains in user orders
				{
					resp, body := testRequest(t, ts, http.MethodGet, userOrdersPath, nil, cookies)
					assert.Contains(
						t,
						resp.Header.Get("content-type"),
						"application/json",
						fmt.Sprintf("bad header %v doesn't containt %s", resp.Header.Get("content-type"), "application/json"),
					)
					assert.Equal(t, resp.StatusCode, http.StatusOK, fmt.Sprintf("bad status code want=200 got %d", resp.StatusCode))

					var orders []*model.Order
					err := json.Unmarshal(body, &orders)
					require.NoError(t, err, fmt.Sprintf("got unexpected error while json unmarshalling: %v", err))
					status := false
					for _, o := range orders {
						if o.Number == tt.o {
							status = true
						}
					}
					require.True(t, status, "found order in registered orders by another user")
				}
				// check doesn't contain in other user
				{
					resp, body := testRequest(t, ts, http.MethodGet, userOrdersPath, nil, otherCookies)
					if resp.StatusCode != http.StatusNoContent {
						assert.Contains(
							t,
							resp.Header["content-type"],
							"application/json",
							fmt.Sprintf("bad header %+v doesn't containt %s", resp.Header["content-type"], "application/json"),
						)
					}
					assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent, fmt.Sprintf("bad status code want=200/204 got %d", resp.StatusCode))

					if resp.StatusCode == http.StatusNoContent {
						return
					}
					var orders []*model.Order
					err := json.Unmarshal(body, &orders)
					require.NoError(t, err, fmt.Sprintf("got unexpected error while json unmarshalling: %v", err))
					status := false
					for _, o := range orders {
						if o.Number == tt.o {
							status = true
						}
					}
					require.False(t, status, "found order in registered orders by another user")
				}
			}
		})
	}

}

func TestWithdrawsPost(t *testing.T) {
	if true {
		return
	}
	if conStr == "" {
		t.Skip("connect string is not provided")
	}
	cfg := config.TestConfig(t)
	ctx := context.Background()

	storage, teardown := sqlstore.TestStore(t, conStr)
	defer teardown(userTableName, ordersTableName)

	log := logrus.New()
	log.Out = io.Discard

	p := poller.TestPoller(t, storage)

	s := server.New(logger.GetLoggerByEntry(logrus.NewEntry(log)), storage, cfg, p)
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	var err error

	u1 := &model.User{Login: userLogin1, Password: userPassword}
	//t.Logf("user id=%d password=%s enc_pass=%s", u1.ID, u1.Password, u1.EncryptedPassword)
	require.NoError(t, err, fmt.Sprintf("got unexpected error while creating user: %v", err))
	u1.Password = userPassword
	cookiesU1 := getUserCookies(t, ts, u1)
	u1, err = storage.User().GetByLogin(ctx, u1.Login)
	require.NoError(t, err, fmt.Sprintf("get user by login: %v", err))

	u2 := &model.User{Login: userLogin2, Password: userPassword}
	require.NoError(t, err, fmt.Sprintf("got unexpected error while creating user: %v", err))
	u2.Password = userPassword
	//cookiesU2 := getUserCookies(t, ts, u2)
	u2, err = storage.User().GetByLogin(ctx, u2.Login)
	require.NoError(t, err, fmt.Sprintf("get user by login: %v", err))

	err = storage.Order().Register(ctx, u1.ID, validOrderNum1)
	require.NoError(t, err, "got unexpected error: %v", err)

	type want struct {
		status int
	}
	type args struct {
		order   int
		sum     float64
		cookies []*http.Cookie
	}

	tests := []struct {
		name string
		args args

		want want
	}{
		{
			name: "positive case #1",
			args: args{
				cookies: cookiesU1,
				order:   validOrderNum1,
				sum:     123.0,
			},
			want: want{
				status: http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			resp, _ := testRequest(t, ts, http.MethodPost, userWithdrawPath, body, tt.args.cookies)
			defer require.NoError(t, resp.Body.Close())
			require.Equal(t, resp.StatusCode, http.StatusOK)
		})
	}

}
