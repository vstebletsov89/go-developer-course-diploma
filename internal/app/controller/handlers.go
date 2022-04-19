package controller

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/theplant/luhn"
	"go-developer-course-diploma/internal/app/configs"
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

func WriteResponseJSON(w http.ResponseWriter, statusCode int, body interface{}) {
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

type Controller struct {
	Config  *configs.Config
	Storage storage.Storage
	Logger  *logrus.Logger
}

func NewController(c *configs.Config, s storage.Storage, l *logrus.Logger) *Controller {
	return &Controller{Config: c, Storage: s, Logger: l}
}

func (c *Controller) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("RegisterHandler: start")
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		c.Logger.Debug("RegisterHandler: check user data")
		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteError(w, http.StatusBadRequest, errors.New("login and password must NOT be empty"))
			return
		}

		getPasswordHash(user)

		c.Logger.Debug("RegisterHandler: RegisterUser")
		err := c.Storage.Users().RegisterUser(user)
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
		c.Logger.Debug("RegisterHandler: end")
	}
}

func (c *Controller) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("LoginHandler: start")
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		c.Logger.Debug("LoginHandler: check user data")
		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteError(w, http.StatusBadRequest, errors.New("login and password must NOT be empty"))
			return
		}

		c.Logger.Debug("LoginHandler: GetUser")
		userDB, err := c.Storage.Users().GetUser(user.Login)
		if err != nil && !errors.Is(err, storage.ErrorUserNotFound) {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorUserNotFound) {
			WriteError(w, http.StatusUnauthorized, err)
			return
		}

		getPasswordHash(user)

		c.Logger.Debug("LoginHandler: user found check credentials")
		if userDB.Login == user.Login && userDB.Password == user.Password {
			auth.SetCookie(w, user.Login)
			WriteResponse(w, http.StatusOK, "")
			return
		}
		WriteResponse(w, http.StatusUnauthorized, "")
		c.Logger.Debug("LoginHandler: end")
	}
}

func (c *Controller) UploadOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("UploadOrder: start")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		number := string(b)
		c.Logger.Debugf("UploadOrder number: %s", number)

		if !IsValidOrderNumber(number) {
			WriteResponse(w, http.StatusUnprocessableEntity, "")
			return
		}

		user, err := auth.GetUser(r)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		userDB, err := c.Storage.Orders().GetUserByOrderNumber(number)
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
			c.Logger.Debug("UploadOrder: REGISTERED")
			err := c.Storage.Orders().UploadOrder(order)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, err)
				return
			}

			if err := c.Storage.Orders().UpdateOrderStatus(order); err != nil {
				WriteError(w, http.StatusInternalServerError, err)
				return
			}

			// try to get loyalty points in goroutine
			var numbers []string
			go c.UpdatePendingOrders(append(numbers, number))

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

func (c *Controller) UpdatePendingOrders(orders []string) error {
	for _, o := range orders {
		link := fmt.Sprintf("%s/api/orders/%s", c.Config.AccrualSystemAddress, o)
		req, err := http.NewRequest(http.MethodGet, link, nil)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		var order *model.Order
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			return err
		}
		defer resp.Body.Close()

		c.Logger.Debugf("Updated order '%s' status '%s' accrual '%d' : \n", order.Number, order.Status, order.Accrual)

		if err := c.Storage.Orders().UpdateOrderStatus(order); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) GetOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("GetOrders: start")
		user, err := auth.GetUser(r)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		c.Logger.Debug("GetOrders: get user orders")
		orders, err := c.Storage.Orders().GetOrders(user)
		if err != nil && !errors.Is(err, storage.ErrorOrderNotFound) {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorOrderNotFound) {
			WriteError(w, http.StatusNoContent, err)
			return
		}
		c.Logger.Debug("GetOrders: write response")
		WriteResponseJSON(w, http.StatusOK, orders)
	}
}
