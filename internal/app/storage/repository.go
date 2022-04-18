package storage

import (
	"errors"
	"go-developer-course-diploma/internal/app/model"
)

var ErrorUserAlreadyExist = errors.New("user already exist")
var ErrorUserNotFound = errors.New("user not found")

type UserRepository interface {
	RegisterUser(*model.User) error
	GetUser(string) (*model.User, error)
}
