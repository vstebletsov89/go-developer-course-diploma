package server

import (
	"github.com/gorilla/mux"
	"go-developer-course-diploma/internal/controller"
	"go-developer-course-diploma/internal/service/auth"
	"net/http"
)

type server struct {
	router *mux.Router
}

func NewServer(controller *controller.Controller) *server {
	s := &server{
		router: mux.NewRouter(),
	}
	s.NewRouter(controller)
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) NewRouter(controller *controller.Controller) {
	controller.Logger.Info("Routing started")
	s.router.HandleFunc("/api/user/register", controller.RegisterHandler()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/user/login", controller.LoginHandler()).Methods(http.MethodPost)

	secure := s.router.NewRoute().Subrouter()
	secure.Use(auth.MiddlewareGeneratorAuthorization(controller.UserAuthorizationStore))
	secure.HandleFunc("/api/user/orders", controller.UploadOrder()).Methods(http.MethodPost)
	secure.HandleFunc("/api/user/orders", controller.GetOrders()).Methods(http.MethodGet)
	secure.HandleFunc("/api/user/balance", controller.GetCurrentBalance()).Methods(http.MethodGet)
	secure.HandleFunc("/api/user/balance/withdraw", controller.WithdrawLoyaltyPoints()).Methods(http.MethodPost)
	secure.HandleFunc("/api/user/balance/withdrawals", controller.GetWithdrawals()).Methods(http.MethodGet)
}
