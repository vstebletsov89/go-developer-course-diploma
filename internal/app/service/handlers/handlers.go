package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/theplant/luhn"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/service/auth"
	"go-developer-course-diploma/internal/app/storage"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	REGISTERED = "REGISTERED"
	INVALID    = "INVALID"
	PROCESSING = "PROCESSING"
	PROCESSED  = "PROCESSED"
)

func WriteError(w http.ResponseWriter, code int, err error) {
	WriteResponse(w, code, err.Error())
}

func WriteResponse(w http.ResponseWriter, statusCode int, data string) {
	w.WriteHeader(statusCode)
	if len(data) != 0 {
		w.Write([]byte(data))
	}
}

func ReturnJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}
	WriteResponse(w, statusCode, "")
}

func IsValidOrderNumber(number string) bool {
	value, err := strconv.Atoi(number)
	if err != nil {
		return false
	}
	return luhn.Valid(value)
}

func getPasswordHash(u *model.User) {
	u.Password = hex.EncodeToString([]byte(u.Password))
}

func RegisterHandler(s storage.Storage, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("RegisterHandler: start")
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		logger.Info("RegisterHandler: check user data")
		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteError(w, http.StatusBadRequest, errors.New("login and password must NOT be empty"))
			return
		}

		getPasswordHash(user)

		logger.Info("RegisterHandler: RegisterUser")
		err := s.Users().RegisterUser(user)
		if err != nil && !errors.Is(err, storage.ErrorUserAlreadyExist) {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorUserAlreadyExist) {
			WriteError(w, http.StatusConflict, err)
			return
		}

		auth.SetCookie(w, user.Login)
		WriteResponse(w, http.StatusOK, "")
		logger.Info("RegisterHandler: end")
	}
}

func LoginHandler(s storage.Storage, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("LoginHandler: start")
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		logger.Info("LoginHandler: check user data")
		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteError(w, http.StatusBadRequest, errors.New("login and password must NOT be empty"))
			return
		}

		logger.Info("LoginHandler: GetUser")
		userDB, err := s.Users().GetUser(user.Login)
		if err != nil && !errors.Is(err, storage.ErrorUserNotFound) {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorUserNotFound) {
			WriteError(w, http.StatusUnauthorized, err)
			return
		}

		getPasswordHash(user)

		logger.Info("LoginHandler: user found check credentials")
		if userDB.Login == user.Login && userDB.Password == user.Password {
			auth.SetCookie(w, user.Login)
			WriteResponse(w, http.StatusOK, "")
			return
		}
		WriteResponse(w, http.StatusUnauthorized, "")
		logger.Info("LoginHandler: end")
	}
}

func UploadOrder(s storage.Storage, accrualSystemAddress string, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		number := string(b)

		if !IsValidOrderNumber(number) {
			WriteResponse(w, http.StatusUnprocessableEntity, "")
			return
		}

		user, err := auth.GetUser(r)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		userDB, err := s.Orders().GetUserByOrderNumber(number)
		if err != nil && !errors.Is(err, storage.ErrorOrderNotFound) {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorOrderNotFound) {
			order := &model.Order{
				Number: number,
				Status: REGISTERED,
				Login:  user,
			}

			err := s.Orders().UploadOrder(order)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, err)
				return
			}

			//TODO: add attempts to get loyalty points?
			go getLoyaltyPoints(s, accrualSystemAddress, number)

			WriteResponse(w, http.StatusAccepted, "")
			return
		}

		if userDB == user {
			WriteResponse(w, http.StatusOK, "")
			return
		}
		// process error when same order number was processed by another user
		WriteResponse(w, http.StatusConflict, "")
	}
}

func getLoyaltyPoints(s storage.Storage, accrualSystemAddress string, orderNumber string) {
	link := fmt.Sprintf("%s/api/orders/%s", accrualSystemAddress, orderNumber)
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	var order *model.Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return
	}
	defer resp.Body.Close()

	if err := s.Orders().UpdateOrderStatus(order); err != nil {
		return
	}
}

func GetOrders(s storage.Storage, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		orders, err := s.Orders().GetOrders(user)
		if err != nil && !errors.Is(err, storage.ErrorOrderNotFound) {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorOrderNotFound) {
			WriteError(w, http.StatusNoContent, err)
			return
		}

		ReturnJSON(w, http.StatusOK, orders)
	}
}
