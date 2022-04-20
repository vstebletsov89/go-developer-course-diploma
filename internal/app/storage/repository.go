package storage

import (
	"errors"
	"go-developer-course-diploma/internal/app/model"
)

var ErrorUnauthorized = errors.New("user is unauthorized")
var ErrorUserAlreadyExist = errors.New("user already exist")
var ErrorUserNotFound = errors.New("user not found")
var ErrorOrderNotFound = errors.New("order not found")
var ErrorWithdrawalNotFound = errors.New("withdrawal not found")

type UserRepository interface {
	RegisterUser(*model.User) error
	GetUser(string) (*model.User, error)
}

type OrderRepository interface {
	UploadOrder(*model.Order) error
	GetOrders(string) ([]*model.Order, error)
	GetUserByOrderNumber(string) (string, error)
	UpdateOrderStatus(*model.Order) error
	GetPendingOrders() ([]string, error)
}

type WithdrawRepository interface {
	Withdraw(*model.Withdraw) error
	GetCurrentBalance(string) (float64, error)
	GetWithdrawnAmount(string) (float64, error)
	GetWithdrawals(string) ([]*model.Withdraw, error)
}
