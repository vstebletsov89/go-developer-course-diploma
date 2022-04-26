package auth

import (
	"github.com/google/uuid"
	"go-developer-course-diploma/internal/app/service/auth/secure"
	"go-developer-course-diploma/internal/app/storage/repository"
	"net/http"
	"sync"
	"time"
)

type UserContextType int64

const (
	cookieName                 = "gophermart"
	UserIDCtx  UserContextType = 0
)

type Session struct {
	UserID    int64
	ExpiredAt time.Time
}

type UserAuthorizationStore struct {
	sessions map[string]Session
	mu       sync.RWMutex
}

func NewUserAuthorizationStore() *UserAuthorizationStore {
	return &UserAuthorizationStore{sessions: make(map[string]Session)}
}

var _ secure.UserAuthorization = (*UserAuthorizationStore)(nil)

func (s *UserAuthorizationStore) SetCookie(w http.ResponseWriter, userID int64) {
	sessionID := uuid.NewString()
	s.mu.Lock()
	s.sessions[sessionID] = Session{
		UserID:    userID,
		ExpiredAt: time.Now().Add(time.Hour * 24),
	}
	s.mu.Unlock()

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

func (s *UserAuthorizationStore) GetUserID(r *http.Request) (int64, error) {
	if s.IsValidAuthorization(r) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return 0, err
		}
		return s.sessions[cookie.Value].UserID, nil
	}
	return 0, repository.ErrorUnauthorized
}
