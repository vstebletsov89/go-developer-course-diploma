package accrual

import (
	"encoding/json"
	"fmt"
	"go-developer-course-diploma/internal/model"
	"net/http"
)

type Provider struct {
	accrualSystemAddress string
}

func NewAccrualProvider(accrualAddress string) *Provider {
	return &Provider{
		accrualSystemAddress: accrualAddress,
	}
}

func (p *Provider) GetOrder(orderNumber string) (*model.Order, error) {
	link := fmt.Sprintf("%s/api/orders/%s", p.accrualSystemAddress, orderNumber)
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var order *model.Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return order, nil
}
