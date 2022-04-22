package auth

import (
	"go-developer-course-diploma/internal/app/service/auth/secure"
	"go-developer-course-diploma/internal/app/storage/repository"
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

type UserAuthorizationStore struct {
	sessions map[string]Session
}

func NewUserAuthorizationStore() *UserAuthorizationStore {
	return &UserAuthorizationStore{sessions: make(map[string]Session)}
}

var _ secure.UserAuthorization = (*UserAuthorizationStore)(nil)

func (s *UserAuthorizationStore) SetCookie(w http.ResponseWriter, login string) {
	sessionID := generateRandomString()
	s.sessions[sessionID] = Session{
		Login:     login,
		ExpiredAt: time.Now().Add(time.Hour * 24),
	}

	cookie := &http.Cookie{
		Name:  cookieName,
		Value: sessionID,
	}
	http.SetCookie(w, cookie)
}

func (s *UserAuthorizationStore) IsValidAuthorization(r *http.Request) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return false
	}
	if s.sessions[cookie.Value].ExpiredAt.Before(time.Now()) {
		return false
	}
	return true
}

func (s *UserAuthorizationStore) GetUser(r *http.Request) (string, error) {
	if s.IsValidAuthorization(r) {
		cookie, _ := r.Cookie(cookieName)
		return s.sessions[cookie.Value].Login, nil
	}
	return "", repository.ErrorUnauthorized
}

func generateRandomString() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 20)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
