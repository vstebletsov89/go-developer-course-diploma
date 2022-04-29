package auth

import (
	"go-developer-course-diploma/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(pwd string) (string, error) {
	// use GenerateFromPassword to hash & salt pwd.
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func IsUserAuthorized(user *model.User, userDB *model.User) (bool, error) {
	if user.Login != userDB.Login {
		return false, nil
	}
	err := bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(user.Password))
	if err != nil {
		return false, err
	}

	// user authorized
	return true, nil
}
