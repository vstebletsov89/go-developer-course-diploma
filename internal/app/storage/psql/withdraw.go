package psql

import (
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage"
)

type WithdrawRepository struct {
	Storage *Storage
}

func (r *WithdrawRepository) Withdraw(o *model.Withdraw) error {
	err := r.Storage.DB.QueryRow(
		"INSERT INTO operations (login, order, amount, processed_at) VALUES ($1, $2, $3, NOW()) RETURNING id",
		o.Login,
		o.Order,
		o.Amount,
	).Scan(&o.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *WithdrawRepository) GetCurrentBalance(login string) (float64, error) {
	var balance *float64
	err := r.Storage.DB.QueryRow(
		"SELECT sum(amount) from operations where owner = $1",
		login,
	).Scan(&balance)

	if err != nil {
		return 0, err
	}

	return *balance, nil
}

func (r *WithdrawRepository) GetWithdrawnAmount(login string) (float64, error) {
	var count *float64
	err := r.Storage.DB.QueryRow(
		"SELECT count(amount) from operations where owner = $1 AND amount < 0",
		login,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return *count, nil
}

func (r *WithdrawRepository) GetWithdrawals(login string) ([]*model.Withdraw, error) {
	var withdrawals []*model.Withdraw

	rows, err := r.Storage.DB.Query(
		"SELECT order, amount, processed_at FROM operations WHERE owner = $1",
		login,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		o := &model.Withdraw{}
		err := rows.Scan(
			&o.Order,
			&o.Amount,
			&o.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, storage.ErrorWithdrawalNotFound
	}

	return withdrawals, nil
}
