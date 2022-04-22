package controller

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go-developer-course-diploma/internal/app/configs"
	"go-developer-course-diploma/internal/app/storage/repository"
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

	logger := logrus.New()
	c := NewController(&configs.Config{RunAddress: "localhost:8080", AccrualSystemAddress: "localhost:8080"}, logger, userStore, orderStore, transactionStore)

	s.NewTestRouter(c)
	return s
}

func (s *server) NewTestRouter(controller *Controller) {
	controller.Logger.Info("Routing started")
	s.router.HandleFunc("/api/user/register", controller.RegisterHandler()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/user/login", controller.LoginHandler()).Methods(http.MethodPost)

	secure := s.router.NewRoute().Subrouter()
	secure.Use(AuthHandleMock)
	secure.HandleFunc("/api/user/orders", controller.UploadOrder()).Methods(http.MethodPost)
	secure.HandleFunc("/api/user/orders", controller.GetOrders()).Methods(http.MethodGet)
	secure.HandleFunc("/api/user/balance", controller.GetCurrentBalance()).Methods(http.MethodGet)
	secure.HandleFunc("/api/user/balance/withdraw", controller.WithdrawLoyaltyPoints()).Methods(http.MethodPost)
	secure.HandleFunc("/api/user/balance/withdrawals", controller.GetWithdrawals()).Methods(http.MethodGet)
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
			name:     "check RegisterHandler (invalid input)",
			path:     "api/user/register",
			jsonBody: `{"": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "check RegisterHandler (empty login)",
			path:     "api/user/register",
			jsonBody: `{"login": "","password": "pass"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "check RegisterHandler (empty password)",
			path:     "api/user/register",
			jsonBody: `{"login": "login","password": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "check RegisterHandler (positive test)",
			path:     "api/user/register",
			jsonBody: `{"login": "user","password": "topsecret"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "",
			},
		},
		{
			name:     "check RegisterHandler (user already exists)",
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
			name:     "check LoginHandler (invalid input)",
			path:     "api/user/login",
			jsonBody: `{"": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "check LoginHandler (empty login)",
			path:     "api/user/login",
			jsonBody: `{"login": "","password": "pass"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "check LoginHandler (empty password)",
			path:     "api/user/login",
			jsonBody: `{"login": "login","password": ""}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "",
			},
		},
		{
			name:     "check LoginHandler (positive test)",
			path:     "api/user/login",
			jsonBody: `{"login": "user","password": "topsecret"}`,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusOK,
				responseBody:   "",
			},
		},
		{
			name:     "check LoginHandler (incorrect password)",
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
