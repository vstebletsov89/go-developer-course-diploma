package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/theplant/luhn"
	"go-developer-course-diploma/internal/app/accrual"
	"go-developer-course-diploma/internal/app/configs"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/service/auth"
	"go-developer-course-diploma/internal/app/service/auth/secure"
	"go-developer-course-diploma/internal/app/storage/repository"
	"io/ioutil"
	"net/http"
	"strconv"
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

		encryptedPassword, err := auth.HashAndSalt(user.Password)
		if err != nil {
			c.Logger.Infof("EncryptPassword error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		user.Password = encryptedPassword
		c.Logger.Debugf("RegisterUser %+v\n\n", user)

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

		c.Logger.Infof("RegisterUser userID: '%d'", userID)
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

		c.Logger.Debugf("LoginHandler %+v\n\n", userDB)
		ok, err := auth.IsUserAuthorized(user, userDB)
		if !ok {
			c.Logger.Infof("User unauthorized")
			WriteResponse(w, http.StatusUnauthorized, "")
			return
		}

		if err != nil {
			c.Logger.Infof("IsUserAuthorized error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// set cookie for authorized user
		c.UserAuthorizationStore.SetCookie(w, userDB.ID)
		WriteResponse(w, http.StatusOK, "")
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

		userDB, err := c.OrderRepository.GetUserIDByOrderNumber(number)
		if err != nil && !errors.Is(err, repository.ErrorOrderNotFound) {
			c.Logger.Infof("GetUserByOrderNumber error: %s", err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		userID := c.extractUserID(r)
		c.Logger.Debugf("UploadOrder: userID '%d'", userID)

		if errors.Is(err, repository.ErrorOrderNotFound) {
			order := &model.Order{
				Number: number,
				Status: accrual.New,
				UserID: userID,
			}

			err := c.OrderRepository.UploadOrder(order)
			if err != nil {
				c.Logger.Infof("UploadOrder error: %s", err)
				WriteError(w, http.StatusInternalServerError, err)
				return
			}

			WriteResponse(w, http.StatusAccepted, "")
			return
		}

		if userDB == userID {
			WriteResponse(w, http.StatusOK, "")
			return
		}

		// process error when same order number was processed by another user
		WriteResponse(w, http.StatusConflict, "")
	}
}

func (c *Controller) GetOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Debug("GetOrders handler")

		userID := c.extractUserID(r)
		c.Logger.Debugf("GetOrders userID '%d'", userID)

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
		c.Logger.Debug("GetWithdrawnAmount repository")
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
