package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/theplant/luhn"
	"go-developer-course-diploma/internal/app/configs"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/service/auth"
	"go-developer-course-diploma/internal/app/service/auth/secure"
	"go-developer-course-diploma/internal/app/storage/repository"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	New        = "NEW"
	Processing = "PROCESSING"
	Invalid    = "INVALID"
	Processed  = "PROCESSED"
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

func IsValidOrderNumber(number string) bool {
	value, err := strconv.Atoi(number)
	if err != nil {
		return false
	}
	return luhn.Valid(value)
}

type Controller struct {
	Config                 *configs.Config
	Logger                 *logrus.Logger
	UserRepository         repository.UserRepository
	OrderRepository        repository.OrderRepository
	TransactionRepository  repository.TransactionRepository
	UserAuthorizationStore secure.UserAuthorization
}

func NewController(cfg *configs.Config, logger *logrus.Logger, userStore repository.UserRepository, orderStore repository.OrderRepository, transactionStore repository.TransactionRepository, userAuthorizationStore secure.UserAuthorization) *Controller {
	return &Controller{
		Config:                 cfg,
		Logger:                 logger,
		UserRepository:         userStore,
		OrderRepository:        orderStore,
		TransactionRepository:  transactionStore,
		UserAuthorizationStore: userAuthorizationStore,
	}
}

func (c *Controller) extractUserID(r *http.Request) int64 {
	userID, ok := r.Context().Value(auth.UserIDCtx).(int64)
	if ok {
		c.Logger.Infof("userID (context): '%d'", userID)
		return userID
	}
	return 0
}

func (c *Controller) WriteJSON(w http.ResponseWriter, response interface{}) {
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(response)
	if err != nil {
		c.Logger.Infof("Encoder error: %s", err)
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	c.Logger.Debugf("WriteJSON response: %s", buf.String())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(buf.Bytes())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}
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

		encryptedPassword, err := auth.EncryptPassword(user.Password)
		if err != nil {
			c.Logger.Infof("EncryptPassword error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		user.Password = encryptedPassword

		userID, err := c.UserRepository.RegisterUser(user)
		if errors.Is(err, repository.ErrorUserAlreadyExist) {
			WriteError(w, http.StatusConflict, err)
			return
		}
		if err != nil {
			c.Logger.Infof("RegisterUser error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		c.UserAuthorizationStore.SetCookie(w, userID)
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
		if errors.Is(err, repository.ErrorUserNotFound) {
			WriteError(w, http.StatusUnauthorized, err)
			return
		}
		if err != nil {
			c.Logger.Infof("GetUser error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		decryptedPassword, err := auth.DecryptPassword(userDB.Password)
		if err != nil {
			c.Logger.Infof("DecryptPassword error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if userDB.Login == user.Login && user.Password == decryptedPassword {
			c.UserAuthorizationStore.SetCookie(w, user.ID)
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
			WriteResponse(w, http.StatusUnprocessableEntity, "invalid order number")
			return
		}

		userID, err := c.OrderRepository.GetUserIDByOrderNumber(number)
		if err != nil && !errors.Is(err, repository.ErrorOrderNotFound) {
			c.Logger.Infof("GetUserByOrderNumber error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if errors.Is(err, repository.ErrorOrderNotFound) {
			order := &model.Order{
				Number: number,
				Status: New,
				UserID: userID,
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

		if order.Status == Processed {
			// set order.Number because response from accrual has 'order' field instead of 'number'
			order.Number = o
			c.Logger.Debugf("Updated order '%s' status '%s' accrual '%f' : \n", order.Number, order.Status, order.Accrual)

			if err := c.OrderRepository.UpdateOrderStatus(order); err != nil {
				c.Logger.Infof("UpdateOrderStatus error: %s", err)
				return err
			}

			// get current user and accumulate balance
			userID, err := c.OrderRepository.GetUserIDByOrderNumber(order.Number)
			if err != nil {
				c.Logger.Infof("GetUserIDByOrderNumber error: %s", err)
				return err
			}

			transaction := &model.Transaction{UserID: userID, Order: order.Number, Amount: order.Accrual}

			c.Logger.Debugf("%+v\n", transaction)

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

		userID := c.extractUserID(r)
		response, err := c.OrderRepository.GetOrders(userID)

		if errors.Is(err, repository.ErrorOrderNotFound) {
			c.Logger.Infof("GetOrders error: %s", err)
			WriteError(w, http.StatusNoContent, err)
			return
		}
		if err != nil {
			c.Logger.Infof("GetOrders error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		c.WriteJSON(w, response)
	}
}

func (c *Controller) GetCurrentBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("GetCurrentBalance handler")
		userID := c.extractUserID(r)
		balance, err := c.TransactionRepository.GetCurrentBalance(userID)
		if err != nil {
			c.Logger.Infof("GetCurrentBalance error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		withdrawn, err := c.TransactionRepository.GetWithdrawnAmount(userID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		response := &model.Balance{
			Current:   balance,
			Withdrawn: withdrawn,
		}

		c.WriteJSON(w, response)
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
			WriteResponse(w, http.StatusBadRequest, "withdraw sum should be greater than zero")
			return
		}

		if !IsValidOrderNumber(withdraw.Order) {
			WriteResponse(w, http.StatusUnprocessableEntity, "invalid order number")
			return
		}

		userID := c.extractUserID(r)

		balance, err := c.TransactionRepository.GetCurrentBalance(userID)
		if err != nil {
			c.Logger.Infof("GetCurrentBalance error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if balance < withdraw.Amount {
			WriteResponse(w, http.StatusPaymentRequired, "insufficient loyalty points")
			return
		}

		withdraw.UserID = userID
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
		userID := c.extractUserID(r)

		response, err := c.TransactionRepository.GetWithdrawals(userID)
		if errors.Is(err, repository.ErrorWithdrawalNotFound) {
			c.Logger.Infof("GetWithdrawals error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if err != nil {
			c.Logger.Infof("GetWithdrawals error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		c.WriteJSON(w, response)
	}
}
