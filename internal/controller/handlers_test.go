package controller

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go-developer-course-diploma/internal/configs"
	"go-developer-course-diploma/internal/service/auth/secure"
	"go-developer-course-diploma/internal/storage/repository"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	assert.NoError(t, err)

	client := &http.Client{}

	resp, err := client.Do(req)
	assert.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	return resp, string(respBody)
}

func AuthHandleMock(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

type server struct {
	router *mux.Router
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServerTest() *server {
	s := &server{
		router: mux.NewRouter(),
	}
	userStore := repository.NewMockRepository()
	orderStore := repository.NewMockOrderRepository()
	transactionStore := repository.NewMockTransactionRepository()
	userAuthStore := secure.NewMockUserAuthorizationStore()

	logger := logrus.New()
	c := NewController(&configs.Config{RunAddress: "localhost:8080", AccrualSystemAddress: "localhost:8080"}, logger, userStore, orderStore, transactionStore, userAuthStore)

	s.NewTestRouter(c)
	return s
}

func (s *server) NewTestRouter(controller *Controller) {
	controller.Logger.Info("Routing started")
	s.router.HandleFunc("/api/user/register", controller.RegisterHandler()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/user/login", controller.LoginHandler()).Methods(http.MethodPost)

	subRouter := s.router.NewRoute().Subrouter()
	subRouter.Use(AuthHandleMock)
	subRouter.HandleFunc("/api/user/orders", controller.UploadOrder()).Methods(http.MethodPost)
	subRouter.HandleFunc("/api/user/orders", controller.GetOrders()).Methods(http.MethodGet)
	subRouter.HandleFunc("/api/user/balance", controller.GetCurrentBalance()).Methods(http.MethodGet)
	subRouter.HandleFunc("/api/user/balance/withdraw", controller.WithdrawLoyaltyPoints()).Methods(http.MethodPost)
	subRouter.HandleFunc("/api/user/balance/withdrawals", controller.GetWithdrawals()).Methods(http.MethodGet)
}

func TestGetGetWithdrawals(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "GetWithdrawals (positive test)",
			path: "api/user/balance/withdrawals",
			body: "",
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "[{\"order\":\"10001\",\"sum\":50.6,\"processed_at\":\"0001-01-01T00:00:00Z\"},{\"order\":\"10002\",\"sum\":789.45,\"processed_at\":\"0001-01-01T00:00:00Z\"},{\"order\":\"10003\",\"sum\":256.9812345,\"processed_at\":\"0001-01-01T00:00:00Z\"}]\n",
			},
		},
	}

	srv := NewServerTest()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%s", tt.path), bytes.NewBufferString(tt.body))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}

func TestWithdrawLoyaltyPoints(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "WithdrawLoyaltyPoints (invalid json)",
			path: "api/user/balance/withdraw",
			body: `{{"": "","": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "invalid character '{' looking for beginning of object key string",
			},
		},
		{
			name: "WithdrawLoyaltyPoints (withdraw.Amount < 0)",
			path: "api/user/balance/withdraw",
			body: `{"order": "2377225624","sum": -751.24}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "withdraw sum should be greater than zero",
			},
		},
		{
			name: "WithdrawLoyaltyPoints (invalid order number)",
			path: "api/user/balance/withdraw",
			body: `{"order": "aaa","sum": 751.24}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusUnprocessableEntity,
				responseBody:   "invalid order number",
			},
		},
		{
			name: "WithdrawLoyaltyPoints (balance < withdraw.Amount)",
			path: "api/user/balance/withdraw",
			body: `{"order": "2377225624","sum": 10000}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusPaymentRequired,
				responseBody:   "insufficient loyalty points",
			},
		},
		{
			name: "WithdrawLoyaltyPoints (positive test)",
			path: "api/user/balance/withdraw",
			body: `{"order": "2377225624","sum": 5000}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "",
			},
		},
	}

	srv := NewServerTest()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/%s", tt.path), bytes.NewBufferString(tt.body))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}

func TestGetCurrentBalance(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "GetCurrentBalance (positive test)",
			path: "api/user/balance",
			body: "",
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "{\"current\":9000.456,\"withdrawn\":3000.15}\n",
			},
		},
	}

	srv := NewServerTest()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%s", tt.path), bytes.NewBufferString(tt.body))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}

func TestGetOrders(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "GetOrders (positive test)",
			path: "api/user/orders",
			body: "",
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "[{\"number\":\"10001\",\"status\":\"PROCESSED\",\"uploaded_at\":\"0001-01-01T00:00:00Z\"},{\"number\":\"10002\",\"status\":\"PROCESSING\",\"uploaded_at\":\"0001-01-01T00:00:00Z\"},{\"number\":\"10003\",\"status\":\"NEW\",\"uploaded_at\":\"0001-01-01T00:00:00Z\"}]\n",
			},
		},
	}

	srv := NewServerTest()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%s", tt.path), bytes.NewBufferString(tt.body))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}

func TestUploadOrder(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "UploadOrder (incorrect order number)",
			path: "api/user/orders",
			body: "aaa",
			want: want{
				headerLocation: "",
				statusCode:     http.StatusUnprocessableEntity,
				responseBody:   "invalid order number",
			},
		},
		{
			name: "UploadOrder (positive test)",
			path: "api/user/orders",
			body: "401869",
			want: want{
				headerLocation: "",
				statusCode:     http.StatusAccepted,
				responseBody:   "",
			},
		},
	}

	srv := NewServerTest()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/%s", tt.path), bytes.NewBufferString(tt.body))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name     string
		path     string
		jsonBody string
		want     want
	}{
		{
			name:     "RegisterHandler (invalid json)",
			path:     "api/user/register",
			jsonBody: `{{"": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "invalid character '{' looking for beginning of object key string",
			},
		},
		{
			name:     "RegisterHandler (empty login)",
			path:     "api/user/register",
			jsonBody: `{"login": "","password": "pass"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "RegisterHandler (empty password)",
			path:     "api/user/register",
			jsonBody: `{"login": "login","password": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "RegisterHandler (positive test)",
			path:     "api/user/register",
			jsonBody: `{"login": "user","password": "topsecret"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "",
			},
		},
		{
			name:     "RegisterHandler (user already exists)",
			path:     "api/user/register",
			jsonBody: `{"login": "user","password": "topsecret"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusConflict,
				responseBody:   "user already exist",
			},
		},
	}

	srv := NewServerTest()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/%s", tt.path), bytes.NewBufferString(tt.jsonBody))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}

func TestLoginHandler(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name     string
		path     string
		jsonBody string
		want     want
	}{
		{
			name:     "LoginHandler (invalid json)",
			path:     "api/user/login",
			jsonBody: `{{"": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "invalid character '{' looking for beginning of object key string",
			},
		},
		{
			name:     "LoginHandler (empty login)",
			path:     "api/user/login",
			jsonBody: `{"login": "","password": "pass"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "LoginHandler (empty password)",
			path:     "api/user/login",
			jsonBody: `{"login": "login","password": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "LoginHandler (positive test)",
			path:     "api/user/login",
			jsonBody: `{"login": "user","password": "topsecret"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "",
			},
		},
		{
			name:     "LoginHandler (incorrect password)",
			path:     "api/user/login",
			jsonBody: `{"login": "user","password": "wronpass"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusUnauthorized,
				responseBody:   "",
			},
		},
	}

	srv := NewServerTest()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// register user
	r, _ := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/%s", "api/user/register"), bytes.NewBufferString(`{"login": "user","password": "topsecret"}`))
	defer r.Body.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/%s", tt.path), bytes.NewBufferString(tt.jsonBody))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}
