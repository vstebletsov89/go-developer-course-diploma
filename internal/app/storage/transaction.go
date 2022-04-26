package storage

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage/repository"
)

type TransactionRepository struct {
	conn *sql.DB
}

func NewTransactionRepository(conn *sql.DB) *TransactionRepository {
	return &TransactionRepository{conn: conn}
}

func (r *TransactionRepository) ExecuteTransaction(t *model.Transaction) error {
	err := r.conn.QueryRow(
		"INSERT INTO transactions (user_id, number, amount, processed_at) VALUES ($1, $2, $3, NOW()) RETURNING id",
		t.UserID,
		t.Order,
		t.Amount,
	).Scan(&t.ID)

	if err != nil {
		return err
	}
	return nil
}

func (r *TransactionRepository) GetCurrentBalance(userID int64) (float64, error) {
	var balance *float64

	err := r.conn.QueryRow(
		"SELECT sum(amount) from transactions where user_id = $1",
		userID,
	).Scan(&balance)

	if err != nil {
		return 0, err
	}

	return *balance, nil
}

func (r *TransactionRepository) GetWithdrawnAmount(userID int64) (float64, error) {
	var count *float64
	err := r.conn.QueryRow(
		"SELECT count(amount) from transactions where user_id = $1 AND amount < 0",
		userID,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	var amount *float64
	if *count > 0 {
		err := r.conn.QueryRow(
			"SELECT sum(amount) from transactions where user_id = $1 AND amount < 0",
			userID,
		).Scan(&amount)

		if err != nil {
			return 0, err
		}
	} else {
		return 0, nil
	}

	return *amount * -1, nil
}

func (r *TransactionRepository) GetWithdrawals(userID int64) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	rows, err := r.conn.Query(
		"SELECT number, amount, processed_at FROM transactions WHERE user_id = $1",
		userID,
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
		return nil, repository.ErrorWithdrawalNotFound
	}

	return transactions, nil
}
