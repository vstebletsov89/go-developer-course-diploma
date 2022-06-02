package auth

import (
	"github.com/google/uuid"
	"go-developer-course-diploma/internal/service/auth/secure"
	"go-developer-course-diploma/internal/storage/repository"
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

func (s *UserAuthorizationStore) storeAuthorization(sessionID string, userID int64) {
	s.mu.Lock()
	s.sessions[sessionID] = Session{
		UserID:    userID,
		ExpiredAt: time.Now().Add(time.Hour * 24),
	}
	s.mu.Unlock()
}

func (s *UserAuthorizationStore) loadAuthorization(sessionID string) (Session, bool) {
	s.mu.RLock()
	session, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	return session, ok
}

func (s *UserAuthorizationStore) SetCookie(w http.ResponseWriter, userID int64) {
	sessionID := uuid.NewString()
	s.storeAuthorization(sessionID, userID)

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
	session, ok := s.loadAuthorization(cookie.Value)
	if !ok || session.ExpiredAt.Before(time.Now()) {
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
		// if authorization is valid then session exists
		session, _ := s.loadAuthorization(cookie.Value)
		return session.UserID, nil
	}
	return 0, repository.ErrorUnauthorized
}
