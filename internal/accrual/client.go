package accrual

import (
	"context"
	"github.com/sirupsen/logrus"
	"go-developer-course-diploma/internal/configs"
	"go-developer-course-diploma/internal/model"
	"go-developer-course-diploma/internal/storage/repository"
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
	accrualProvider       *Provider
	logger                *logrus.Logger
	orderRepository       repository.OrderRepository
	transactionRepository repository.TransactionRepository
}

func NewAccrualClient(cfg *configs.Config, logger *logrus.Logger, orderStore repository.OrderRepository, transactionStore repository.TransactionRepository) *Client {
	return &Client{
		accrualProvider:       NewAccrualProvider(cfg.AccrualSystemAddress),
		logger:                logger,
		orderRepository:       orderStore,
		transactionRepository: transactionStore,
	}
}

func (c *Client) UpdatePendingOrders(orders []string) error {
	c.logger.Debug("UpdatePendingOrders: start")
	for _, o := range orders {

		order, err := c.accrualProvider.GetOrder(o)
		if err != nil {
			c.logger.Infof("GetOrder error: %s", err)
			return err
		}

		if order.Status == Processed {
			// set order.Number because response from accrual has 'order' field instead of 'number'
			order.Number = o
			c.logger.Debugf("Updated order '%s' status '%s' accrual '%f' : \n", order.Number, order.Status, order.Accrual)

			if err := c.orderRepository.UpdateOrderStatus(order); err != nil {
				c.logger.Infof("UpdateOrderStatus error: %s", err)
				return err
			}

			// get current user and accumulate balance
			userID, err := c.orderRepository.GetUserIDByOrderNumber(order.Number)
			if err != nil {
				c.logger.Infof("GetUserIDByOrderNumber error: %s", err)
				return err
			}

			c.logger.Debugf("GetUserIDByOrderNumber userID '%d'", userID)

			transaction := &model.Transaction{UserID: userID, Order: order.Number, Amount: order.Accrual}

			c.logger.Debugf("%+v\n", transaction)

			if err := c.transactionRepository.ExecuteTransaction(transaction); err != nil {
				c.logger.Infof("ExecuteTransaction error: %s", err)
				return err
			}
		}
	}

	c.logger.Debug("UpdatePendingOrders: end")
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
			c.logger.Debug("Check pending orders")
			orders, err := c.orderRepository.GetPendingOrders()
			if err != nil {
				c.logger.Debugf("GetPendingOrders error: %s", err)
			}
			if len(orders) > 0 {
				c.logger.Debug("Update pending orders")
				err := c.UpdatePendingOrders(orders)
				if err != nil {
					c.logger.Debugf("UpdatePendingOrders error: %s", err)
				}
			}
		}
	}
}
