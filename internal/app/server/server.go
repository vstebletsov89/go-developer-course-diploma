package server

import (
	"github.com/gorilla/mux"
	"go-developer-course-diploma/internal/app/service/user"
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

func (s *server) NewRouter(accuralSystemAddress string) {
	s.router.HandleFunc("/api/user/register", user.RegisterHandler(s.storage)).Methods(http.MethodPost)
	s.router.HandleFunc("/api/user/login", user.LoginHandler(s.storage)).Methods(http.MethodPost)

	secure := s.router.NewRoute().Subrouter()
	secure.Use(user.Authorization)
	//TODO: add other endpoints
}
