package storage

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage/repository"
)

type OrderRepository struct {
	conn *sql.DB
}

func NewOrderRepository(conn *sql.DB) *OrderRepository {
	return &OrderRepository{conn: conn}
}

func (r *OrderRepository) UploadOrder(o *model.Order) error {
	err := r.conn.QueryRow(
		"INSERT INTO orders (number, status, user_id, uploaded_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING RETURNING id",
		o.Number,
		o.Status,
		o.UserID,
	).Scan(&o.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) GetUserIDByOrderNumber(number string) (int64, error) {
	var user *int64
	err := r.conn.QueryRow(
		"SELECT user_id FROM orders WHERE number = $1",
		number,
	).Scan(
		&user,
	)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		return 0, repository.ErrorOrderNotFound
	}

	return *user, nil
}

func (r *OrderRepository) GetPendingOrders() ([]string, error) {
	var orders []string

	rows, err := r.conn.Query(
		"SELECT number FROM orders WHERE status = 'NEW' OR status = 'PROCESSING' ORDER BY uploaded_at",
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

func (r *OrderRepository) GetOrders(userID int64) ([]*model.Order, error) {
	var orders []*model.Order

	rows, err := r.conn.Query(
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at",
		userID,
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
		return nil, repository.ErrorOrderNotFound
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(o *model.Order) error {
	err := r.conn.QueryRow(
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		o.Status,
		o.Accrual,
		o.Number,
	).Scan(&o.ID)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}
