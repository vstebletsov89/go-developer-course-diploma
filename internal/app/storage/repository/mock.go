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
}

var _ OrderRepository = (*MockOrderRepository)(nil)

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{}
}

func (m *MockOrderRepository) UploadOrder(order *model.Order) error {
	// do nothing
	return nil
}

func (m *MockOrderRepository) GetOrders(s string) ([]*model.Order, error) {
	var orders []*model.Order
	orders = append(orders, &model.Order{Number: "10001", Status: "PROCESSED", Accrual: 0})
	orders = append(orders, &model.Order{Number: "10002", Status: "PROCESSING", Accrual: 0})
	orders = append(orders, &model.Order{Number: "10003", Status: "NEW", Accrual: 0})
	return orders, nil
}

func (m *MockOrderRepository) GetUserByOrderNumber(s string) (string, error) {
	return s, ErrorOrderNotFound
}

func (m *MockOrderRepository) UpdateOrderStatus(order *model.Order) error {
	// do nothing
	return nil
}

func (m *MockOrderRepository) GetPendingOrders() ([]string, error) {
	// do nothing
	return nil, nil
}

type MockTransactionRepository struct {
	mock.Mock
}

var _ TransactionRepository = (*MockTransactionRepository)(nil)

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{}
}

func (m *MockTransactionRepository) ExecuteTransaction(transaction *model.Transaction) error {
	// do nothing
	return nil
}

func (m *MockTransactionRepository) GetCurrentBalance(s string) (float64, error) {
	// hardcoded balance for tests
	return 9000.456, nil
}

func (m *MockTransactionRepository) GetWithdrawnAmount(s string) (float64, error) {
	// hardcoded withdrawn amount for tests
	return 3000.15, nil
}

func (m *MockTransactionRepository) GetWithdrawals(s string) ([]*model.Transaction, error) {
	var withdrawals []*model.Transaction
	withdrawals = append(withdrawals, &model.Transaction{Order: "10001", Amount: 50.6})
	withdrawals = append(withdrawals, &model.Transaction{Order: "10002", Amount: 789.45})
	withdrawals = append(withdrawals, &model.Transaction{Order: "10003", Amount: 256.9812345})
	return withdrawals, nil
}
