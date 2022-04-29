package secure

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type MockUserAuthorizationStore struct {
	mock.Mock
}

var _ UserAuthorization = (*MockUserAuthorizationStore)(nil)

func NewMockUserAuthorizationStore() *MockUserAuthorizationStore {
	return &MockUserAuthorizationStore{}
}

func (m *MockUserAuthorizationStore) SetCookie(w http.ResponseWriter, userID int64) {
	// do nothing
}

func (m *MockUserAuthorizationStore) IsValidAuthorization(r *http.Request) bool {
	return true
}

func (m *MockUserAuthorizationStore) GetUserID(r *http.Request) (int64, error) {
	// return hardcoded userID for tests
	return 999, nil
}
