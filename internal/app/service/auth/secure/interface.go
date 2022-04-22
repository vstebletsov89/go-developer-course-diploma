package secure

import (
	"net/http"
)

type UserAuthorization interface {
	SetCookie(w http.ResponseWriter, login string)
	IsValidAuthorization(r *http.Request) bool
	GetUser(r *http.Request) (string, error)
}
