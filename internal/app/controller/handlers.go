package controller

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/theplant/luhn"
	"go-developer-course-diploma/internal/app/configs"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/service/auth"
	"go-developer-course-diploma/internal/app/storage"
	"go-developer-course-diploma/internal/app/storage/psql"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	NEW        = "NEW"
	PROCESSING = "PROCESSING"
	INVALID    = "INVALID"
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

func WriteJSON(w http.ResponseWriter, response interface{}, logger *logrus.Logger) {
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(response)
	if err != nil {
		logger.Infof("Encoder error: %s", err)
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	logger.Debugf("WriteJSON response: %s", buf.String())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(buf.Bytes())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}
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
	Config                *configs.Config
	Logger                *logrus.Logger
	UserRepository        *psql.UserRepository
	OrderRepository       *psql.OrderRepository
	TransactionRepository *psql.TransactionRepository
}

func NewController(c *configs.Config, db *sql.DB, l *logrus.Logger) *Controller {
	return &Controller{Config: c, Logger: l,
		UserRepository:        &psql.UserRepository{Conn: db},
		OrderRepository:       &psql.OrderRepository{Conn: db},
		TransactionRepository: &psql.TransactionRepository{Conn: db}}
}

func (c *Controller) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteResponse(w, http.StatusBadRequest, "")
			return
		}

		getPasswordHash(user)

		err := c.UserRepository.RegisterUser(user)
		if err != nil && !errors.Is(err, storage.ErrorUserAlreadyExist) {
			c.Logger.Infof("RegisterUser error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorUserAlreadyExist) {
			WriteError(w, http.StatusConflict, err)
			return
		}

		auth.SetCookie(w, user.Login)
		WriteResponse(w, http.StatusOK, "")
	}
}

func (c *Controller) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteResponse(w, http.StatusBadRequest, "")
			return
		}

		userDB, err := c.UserRepository.GetUser(user.Login)
		if err != nil && !errors.Is(err, storage.ErrorUserNotFound) {
			c.Logger.Infof("GetUser error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorUserNotFound) {
			WriteError(w, http.StatusUnauthorized, err)
			return
		}

		getPasswordHash(user)

		if userDB.Login == user.Login && userDB.Password == user.Password {
			auth.SetCookie(w, user.Login)
			WriteResponse(w, http.StatusOK, "")
			return
		}

		WriteResponse(w, http.StatusUnauthorized, "")
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

		userDB, err := c.OrderRepository.GetUserByOrderNumber(number)
		if err != nil && !errors.Is(err, storage.ErrorOrderNotFound) {
			c.Logger.Infof("GetUserByOrderNumber error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if errors.Is(err, storage.ErrorOrderNotFound) {
			order := &model.Order{
				Number: number,
				Status: NEW,
				Login:  user,
			}

			err := c.OrderRepository.UploadOrder(order)
			if err != nil {
				c.Logger.Infof("UploadOrder error: %s", err)
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
	c.Logger.Debug("UpdatePendingOrders: start")
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

		if order.Status == PROCESSED {
			// set order.Number because response from accrual has 'order' field instead of 'number'
			order.Number = o
			c.Logger.Debugf("Updated order '%s' status '%s' accrual '%f' : \n", order.Number, order.Status, order.Accrual)

			if err := c.OrderRepository.UpdateOrderStatus(order); err != nil {
				c.Logger.Infof("UpdateOrderStatus error: %s", err)
				return err
			}

			// get current user and accumulate balance
			userDB, err := c.OrderRepository.GetUserByOrderNumber(order.Number)
			if err != nil {
				c.Logger.Infof("GetUserByOrderNumber error: %s", err)
				return err
			}

			transaction := &model.Transaction{Login: userDB, Order: order.Number, Amount: order.Accrual}

			log.Printf("%+v\n", transaction)

			if err := c.TransactionRepository.ExecuteTransaction(transaction); err != nil {
				c.Logger.Infof("ExecuteTransaction error: %s", err)
				return err
			}
		}
	}

	c.Logger.Debug("UpdatePendingOrders: end")
	return nil
}

func (c *Controller) GetOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("GetOrders handler")
		user, err := auth.GetUser(r)
		if err != nil {
			c.Logger.Infof("GetUser error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		response, err := c.OrderRepository.GetOrders(user)
		if err != nil && !errors.Is(err, storage.ErrorOrderNotFound) {
			c.Logger.Infof("GetOrders error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorOrderNotFound) {
			c.Logger.Infof("GetOrders error: %s", err)
			WriteError(w, http.StatusNoContent, err)
			return
		}

		WriteJSON(w, response, c.Logger)
	}
}

func (c *Controller) GetCurrentBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("GetCurrentBalance handler")
		user, err := auth.GetUser(r)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		balance, err := c.TransactionRepository.GetCurrentBalance(user)
		if err != nil {
			c.Logger.Infof("GetCurrentBalance error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		withdrawn, err := c.TransactionRepository.GetWithdrawnAmount(user)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		response := &model.Balance{
			Current:   balance,
			Withdrawn: withdrawn,
		}

		WriteJSON(w, response, c.Logger)
	}
}

func (c *Controller) WithdrawLoyaltyPoints() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("WithdrawLoyaltyPoints handler")
		var withdraw *model.Transaction
		if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		if withdraw.Amount < 0 {
			WriteResponse(w, http.StatusBadRequest, "")
			return
		}

		if !IsValidOrderNumber(withdraw.Order) {
			WriteResponse(w, http.StatusUnprocessableEntity, "")
			return
		}

		user, err := auth.GetUser(r)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		balance, err := c.TransactionRepository.GetCurrentBalance(user)
		if err != nil {
			c.Logger.Infof("GetCurrentBalance error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if balance < withdraw.Amount {
			WriteResponse(w, http.StatusPaymentRequired, "")
			return
		}

		withdraw.Login = user
		withdraw.Amount = -1 * withdraw.Amount

		if err := c.TransactionRepository.ExecuteTransaction(withdraw); err != nil {
			c.Logger.Infof("Withdraw error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		WriteResponse(w, http.StatusOK, "")
	}
}

func (c *Controller) GetWithdrawals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("GetWithdrawals handler")
		user, err := auth.GetUser(r)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		response, err := c.TransactionRepository.GetWithdrawals(user)
		if err != nil && err != storage.ErrorWithdrawalNotFound {
			c.Logger.Infof("GetWithdrawals error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if err == storage.ErrorWithdrawalNotFound {
			c.Logger.Infof("GetWithdrawals error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		WriteJSON(w, response, c.Logger)
	}
}
