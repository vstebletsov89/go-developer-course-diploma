package psql

import (
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage"
	"log"
)

type TransactionRepository struct {
	Storage *Storage
}

func (r *TransactionRepository) Transaction(o *model.Transaction) error {
	err := r.Storage.DB.QueryRow(
		"INSERT INTO transactions (login, number, amount, processed_at) VALUES ($1, $2, $3, NOW()) RETURNING id",
		o.Login,
		o.Order,
		o.Amount,
	).Scan(&o.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *TransactionRepository) GetCurrentBalance(login string) (float64, error) {
	var balance *float64
	log.Print("GetCurrentBalance sql start")
	err := r.Storage.DB.QueryRow(
		"SELECT sum(amount) from transactions where login = $1",
		login,
	).Scan(&balance)

	if err != nil {
		return 0, err
	}

	return *balance, nil
}

func (r *TransactionRepository) GetWithdrawnAmount(login string) (float64, error) {
	var count *float64
	err := r.Storage.DB.QueryRow(
		"SELECT count(amount) from transactions where login = $1 AND amount < 0",
		login,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return *count, nil
}

func (r *TransactionRepository) GetWithdrawals(login string) ([]*model.Transaction, error) {
	var transactions []*model.Transaction

	rows, err := r.Storage.DB.Query(
		"SELECT order, amount, processed_at FROM transactions WHERE login = $1",
		login,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		o := &model.Transaction{}
		err := rows.Scan(
			&o.Order,
			&o.Amount,
			&o.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return nil, storage.ErrorWithdrawalNotFound
	}

	return transactions, nil
}
