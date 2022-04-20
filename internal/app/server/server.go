package server

import (
	"github.com/gorilla/mux"
	"go-developer-course-diploma/internal/app/controller"
	"go-developer-course-diploma/internal/app/service/auth"
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
	// `POST /api/user/register` — регистрация пользователя;
	// `POST /api/user/login` — аутентификация пользователя;
	// `POST /api/user/orders` — загрузка пользователем номера заказа для расчёта;
	// `GET /api/user/orders` — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
	// `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности пользователя;
	// `POST /api/user/balance/withdraw` — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
	// `GET /api/user/balance/withdrawals` — получение информации о выводе средств с накопительного счёта пользователем.

	controller.Logger.Info("Routing started")
	s.router.HandleFunc("/api/user/register", controller.RegisterHandler()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/user/login", controller.LoginHandler()).Methods(http.MethodPost)

	secure := s.router.NewRoute().Subrouter()
	secure.Use(auth.Authorization)
	secure.HandleFunc("/api/user/orders", controller.UploadOrder()).Methods(http.MethodPost)
	secure.HandleFunc("/api/user/orders", controller.GetOrders()).Methods(http.MethodGet)
	secure.HandleFunc("/api/user/balance", controller.GetCurrentBalance()).Methods(http.MethodGet)
	secure.HandleFunc("/api/user/balance/withdraw", controller.WithdrawLoyaltyPoints()).Methods(http.MethodPost)
	secure.HandleFunc("/api/user/balance/withdrawals", controller.GetWithdrawals()).Methods(http.MethodGet)
}
