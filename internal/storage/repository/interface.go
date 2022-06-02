package repository

import (
	"errors"
	"go-developer-course-diploma/internal/model"
)

var ErrorUnauthorized = errors.New("user is unauthorized")
var ErrorUserAlreadyExist = errors.New("user already exist")
var ErrorUserNotFound = errors.New("user not found")
var ErrorOrderNotFound = errors.New("order not found")
var ErrorWithdrawalNotFound = errors.New("withdrawal not found")

type UserRepository interface {
	RegisterUser(*model.User) (int64, error)
	GetUser(string) (*model.User, error)
}

type OrderRepository interface {
	UploadOrder(*model.Order) error
	GetOrders(int64) ([]*model.Order, error)
	GetUserIDByOrderNumber(string) (int64, error)
	UpdateOrderStatus(*model.Order) error
	GetPendingOrders() ([]string, error)
}

type TransactionRepository interface {
	ExecuteTransaction(*model.Transaction) error
	GetCurrentBalance(int64) (float64, error)
	GetWithdrawnAmount(int64) (float64, error)
	GetWithdrawals(int64) ([]*model.Transaction, error)
}
