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
	controller.Logger.Info("Create new server...")
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
	controller.Logger.Info("Routing started...")
	s.router.HandleFunc("/api/user/register", controller.RegisterHandler()).Methods(http.MethodPost)
	s.router.HandleFunc("/api/user/login", controller.LoginHandler()).Methods(http.MethodPost)

	secure := s.router.NewRoute().Subrouter()
	secure.Use(auth.Authorization)
	secure.HandleFunc("/api/user/orders", controller.UploadOrder()).Methods(http.MethodPost)
	secure.HandleFunc("/api/user/orders", controller.GetOrders()).Methods(http.MethodGet)
}
