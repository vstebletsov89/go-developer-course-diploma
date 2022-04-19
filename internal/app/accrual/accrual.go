package accrual

import (
	"context"
	"go-developer-course-diploma/internal/app/controller"
	"time"
)

func UpdatePendingOrders(controller *controller.Controller, ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return

		// check pending orders each second
		case <-time.After(1 * time.Second):
			controller.Logger.Debug("Check pending orders")
			orders, err := controller.Storage.Orders().GetPendingOrders()
			if err != nil {
				controller.Logger.Debug("GetPendingOrders failed")
				controller.Logger.Debugf("GetPendingOrders error: %s", err)
			}
			if len(orders) > 0 {
				controller.Logger.Debugf("Pending orders: %+v\n", orders)
				controller.Logger.Debug("Update pending orders")
				controller.UpdatePendingOrders(orders)
			}
		}
	}
}
