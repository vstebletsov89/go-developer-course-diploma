package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go-developer-course-diploma/internal/configs"
	model2 "go-developer-course-diploma/internal/model"
	"go-developer-course-diploma/internal/storage/repository"
	"net/http"
	"time"
)

const (
	New            = "NEW"
	Processing     = "PROCESSING"
	Invalid        = "INVALID"
	Processed      = "PROCESSED"
	pollingTimeout = 1 * time.Second
)

type Client struct {
	Config                *configs.Config
	Logger                *logrus.Logger
	OrderRepository       repository.OrderRepository
	TransactionRepository repository.TransactionRepository
}

func NewAccrualClient(cfg *configs.Config, logger *logrus.Logger, orderStore repository.OrderRepository, transactionStore repository.TransactionRepository) *Client {
	return &Client{
		Config:                cfg,
		Logger:                logger,
		OrderRepository:       orderStore,
		TransactionRepository: transactionStore,
	}
}

func (c *Client) UpdatePendingOrders(orders []string) error {
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

		var order *model2.Order
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

			c.Logger.Debugf("GetUserIDByOrderNumber userID '%d'", userID)

			transaction := &model2.Transaction{UserID: userID, Order: order.Number, Amount: order.Accrual}

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

func (c *Client) CheckPendingOrders(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return

		// check pending orders each second
		case <-time.After(pollingTimeout):
			c.Logger.Debug("Check pending orders")
			orders, err := c.OrderRepository.GetPendingOrders()
			if err != nil {
				c.Logger.Debugf("GetPendingOrders error: %s", err)
			}
			if len(orders) > 0 {
				c.Logger.Debug("Update pending orders")
				err := c.UpdatePendingOrders(orders)
				if err != nil {
					c.Logger.Debugf("UpdatePendingOrders error: %s", err)
				}
			}
		}
	}
}
