package storage

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage/repository"
	"log"
)

type TransactionRepository struct {
	conn *sql.DB
}

func NewTransactionRepository(conn *sql.DB) *TransactionRepository {
	return &TransactionRepository{conn: conn}
}

func (r *TransactionRepository) ExecuteTransaction(t *model.Transaction) error {
	log.Print("ExecuteTransaction sql: start")
	log.Printf("%+v\n", t)
	err := r.conn.QueryRow(
		"INSERT INTO transactions (login, number, amount, processed_at) VALUES ($1, $2, $3, NOW()) RETURNING id",
		t.Login,
		t.Order,
		t.Amount,
	).Scan(&t.ID)

	if err != nil {
		return err
	}
	log.Print("ExecuteTransaction sql: end")
	return nil
}

func (r *TransactionRepository) GetCurrentBalance(login string) (float64, error) {
	var balance *float64
	log.Print("GetCurrentBalance sql start")
	log.Printf("Login '%s'", login)
	err := r.conn.QueryRow(
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
	err := r.conn.QueryRow(
		"SELECT count(amount) from transactions where login = $1 AND amount < 0",
		login,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	var amount *float64
	if *count > 0 {
		err := r.conn.QueryRow(
			"SELECT sum(amount) from transactions where login = $1 AND amount < 0",
			login,
		).Scan(&amount)

		if err != nil {
			return 0, err
		}
	} else {
		return 0, nil
	}

	return *amount * -1, nil
}

func (r *TransactionRepository) GetWithdrawals(login string) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	rows, err := r.conn.Query(
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
		return nil, repository.ErrorWithdrawalNotFound
	}

	return transactions, nil
}
