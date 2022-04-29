package repository

import (
	"errors"
	model2 "go-developer-course-diploma/internal/model"
)

var ErrorUnauthorized = errors.New("user is unauthorized")
var ErrorUserAlreadyExist = errors.New("user already exist")
var ErrorUserNotFound = errors.New("user not found")
var ErrorOrderNotFound = errors.New("order not found")
var ErrorWithdrawalNotFound = errors.New("withdrawal not found")

type UserRepository interface {
	RegisterUser(*model2.User) (int64, error)
	GetUser(string) (*model2.User, error)
}

type OrderRepository interface {
	UploadOrder(*model2.Order) error
	GetOrders(int64) ([]*model2.Order, error)
	GetUserIDByOrderNumber(string) (int64, error)
	UpdateOrderStatus(*model2.Order) error
	GetPendingOrders() ([]string, error)
}

type TransactionRepository interface {
	ExecuteTransaction(*model2.Transaction) error
	GetCurrentBalance(int64) (float64, error)
	GetWithdrawnAmount(int64) (float64, error)
	GetWithdrawals(int64) ([]*model2.Transaction, error)
}
