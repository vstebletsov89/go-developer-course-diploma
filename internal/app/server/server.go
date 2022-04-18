package server

import (
	"github.com/gorilla/mux"
	"go-developer-course-diploma/internal/app/service/auth"
	"go-developer-course-diploma/internal/app/service/handlers"
	"go-developer-course-diploma/internal/app/storage"
	"net/http"
)

type server struct {
	router  *mux.Router
	storage storage.Storage
}

func NewServer(storage storage.Storage, accrualSystemAddress string) *server {
	s := &server{
		router:  mux.NewRouter(),
		storage: storage,
	}
	s.NewRouter(accrualSystemAddress)

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) NewRouter(accrualSystemAddress string) {
	s.router.HandleFunc("/api/auth/register", handlers.RegisterHandler(s.storage)).Methods(http.MethodPost)
	s.router.HandleFunc("/api/auth/login", handlers.LoginHandler(s.storage)).Methods(http.MethodPost)

	secure := s.router.NewRoute().Subrouter()
	secure.Use(auth.Authorization)
	secure.HandleFunc("/api/auth/orders", handlers.UploadOrder(s.storage, accrualSystemAddress)).Methods(http.MethodPost)
	secure.HandleFunc("/api/auth/orders", handlers.GetOrders(s.storage)).Methods(http.MethodGet)
}
