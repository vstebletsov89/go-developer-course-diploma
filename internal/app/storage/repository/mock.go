package repository

import (
	"github.com/stretchr/testify/mock"
	"go-developer-course-diploma/internal/app/model"
)

type MockUserRepository struct {
	mock.Mock
	inMemoryMockDB map[string]string
}

var _ UserRepository = (*MockUserRepository)(nil)

func NewMockRepository() *MockUserRepository {
	return &MockUserRepository{inMemoryMockDB: make(map[string]string)}
}

func (m *MockUserRepository) RegisterUser(user *model.User) error {
	_, exist := m.inMemoryMockDB[user.Login]
	if exist {
		return ErrorUserAlreadyExist
	}
	m.inMemoryMockDB[user.Login] = user.Password
	return nil
}

func (m *MockUserRepository) GetUser(login string) (*model.User, error) {
	pass, ok := m.inMemoryMockDB[login]
	if !ok {
		return nil, ErrorUserNotFound
	}
	return &model.User{Login: login, Password: pass}, nil
}

type MockOrderRepository struct {
	mock.Mock
	//TODO: add required fields (map structs?)
}

var _ OrderRepository = (*MockOrderRepository)(nil)

func NewMockOrderRepository() *MockOrderRepository {
	//TODO: ...
	return &MockOrderRepository{}
}

func (m *MockOrderRepository) UploadOrder(order *model.Order) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockOrderRepository) GetOrders(s string) ([]*model.Order, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockOrderRepository) GetUserByOrderNumber(s string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockOrderRepository) UpdateOrderStatus(order *model.Order) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockOrderRepository) GetPendingOrders() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

type MockTransactionRepository struct {
	mock.Mock
	//TODO: add required fields (map structs?)
}

var _ TransactionRepository = (*MockTransactionRepository)(nil)

func NewMockTransactionRepository() *MockTransactionRepository {
	//TODO: ...
	return &MockTransactionRepository{}
}

func (m *MockTransactionRepository) ExecuteTransaction(transaction *model.Transaction) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTransactionRepository) GetCurrentBalance(s string) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTransactionRepository) GetWithdrawnAmount(s string) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTransactionRepository) GetWithdrawals(s string) ([]*model.Transaction, error) {
	//TODO implement me
	panic("implement me")
}
