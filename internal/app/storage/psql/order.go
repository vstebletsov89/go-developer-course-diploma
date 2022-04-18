package psql

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage"
)

type OrderRepository struct {
	Storage *Storage
}

func (r *OrderRepository) UploadOrder(o *model.Order) error {
	err := r.Storage.DB.QueryRow(
		"INSERT INTO orders (number, status, login, uploaded_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING RETURNING id",
		o.Number,
		o.Status,
		o.Login,
	).Scan(&o.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) GetUserByOrderNumber(number string) (string, error) {
	var user *string
	err := r.Storage.DB.QueryRow(
		"SELECT login FROM orders WHERE number = $1",
		number,
	).Scan(
		&user,
	)

	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err == sql.ErrNoRows {
		return "", storage.ErrorOrderNotFound
	}

	return *user, nil
}

func (r *OrderRepository) GetOrders(login string) ([]*model.Order, error) {
	var orders []*model.Order

	rows, err := r.Storage.DB.Query(
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE login = $1 ORDER BY uploaded_at",
		login,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		o := &model.Order{}
		err := rows.Scan(
			&o.Number,
			&o.Status,
			&o.Accrual,
			&o.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, storage.ErrorOrderNotFound
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(o *model.Order) error {
	err := r.Storage.DB.QueryRow(
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		o.Status,
		o.Accrual,
		o.Number,
	).Scan(&o.ID)

	if err != nil {
		return err
	}

	return nil
}
