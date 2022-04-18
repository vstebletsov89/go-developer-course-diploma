package auth

import (
	"go-developer-course-diploma/internal/app/storage"
	"math/rand"
	"net/http"
	"time"
)

const (
	cookieName = "gophermart"
)

type Session struct {
	Login     string
	ExpiredAt time.Time
}

var sessions = make(map[string]Session)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func SetCookie(w http.ResponseWriter, login string) {
	sessionID := generateRandomString()
	sessions[sessionID] = Session{
		Login:     login,
		ExpiredAt: time.Now().Add(time.Hour * 24),
	}

	cookie := &http.Cookie{
		Name:  cookieName,
		Value: sessionID,
	}
	http.SetCookie(w, cookie)
}

func IsValidAuthorization(r *http.Request) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return false
	}
	if sessions[cookie.Value].ExpiredAt.Before(time.Now()) {
		return false
	}
	return true
}

func GetUser(r *http.Request) (string, error) {
	if IsValidAuthorization(r) {
		cookie, _ := r.Cookie(cookieName)
		return sessions[cookie.Value].Login, nil
	}
	return "", storage.ErrorUnauthorized
}

func generateRandomString() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 20)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
