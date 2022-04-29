package secure

import (
	"net/http"
)

type UserAuthorization interface {
	SetCookie(w http.ResponseWriter, userID int64)
	IsValidAuthorization(r *http.Request) bool
	GetUserID(r *http.Request) (int64, error)
}
