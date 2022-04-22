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

func (m *MockUserAuthorizationStore) SetCookie(w http.ResponseWriter, login string) {
	// do nothing
}

func (m *MockUserAuthorizationStore) IsValidAuthorization(r *http.Request) bool {
	return true
}

func (m *MockUserAuthorizationStore) GetUser(r *http.Request) (string, error) {
	// return hardcoded user for tests
	return "mockuser", nil
}
