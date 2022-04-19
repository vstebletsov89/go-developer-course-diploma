package psql

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage"
	"log"
)

type OrderRepository struct {
	Storage *Storage
}

func (r *OrderRepository) UploadOrder(o *model.Order) error {
	log.Println("UploadOrder sql: started")
	log.Printf("%+v\n\n", o)

	err := r.Storage.DB.QueryRow(
		"INSERT INTO orders (number, status, login, uploaded_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING RETURNING id",
		o.Number,
		o.Status,
		o.Login,
	).Scan(&o.ID)

	if err != nil {
		return err
	}

	log.Println("UploadOrder sql: done")
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

func (r *OrderRepository) GetPendingOrders() ([]string, error) {
	var orders []string

	rows, err := r.Storage.DB.Query(
		"SELECT number FROM orders WHERE status = 'PROCESSING' ORDER BY uploaded_at",
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var number *string
		err := rows.Scan(
			&number,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *number)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) GetOrders(login string) ([]*model.Order, error) {
	var orders []*model.Order
	log.Println("GetOrders sql: started")
	rows, err := r.Storage.DB.Query(
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE login = $1 ORDER BY uploaded_at",
		login,
	)

	log.Printf("%+v\n\n", rows)

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
		log.Printf("%+v\n\n", o)
		orders = append(orders, o)
	}

	log.Printf("%+v\n\n", orders)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, storage.ErrorOrderNotFound
	}

	log.Println("GetOrders sql: done")
	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(o *model.Order) error {
	log.Println("UpdateOrderStatus sql: started")
	log.Printf("%+v\n\n", o)
	err := r.Storage.DB.QueryRow(
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		o.Status,
		o.Accrual,
		o.Number,
	).Scan(&o.ID)

	if err != nil {
		return err
	}
	log.Println("UpdateOrderStatus sql: done")
	return nil
}
