package psql

import (
	"database/sql"
	_ "github.com/lib/pq"
	"go-developer-course-diploma/internal/app/storage"
)

type Storage struct {
	DB               *sql.DB
	userRepository   *UserRepository
	ordersRepository *OrderRepository
	//withdrawRepository *WithdrawRepository
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		DB: db,
	}
}

func (s *Storage) Users() storage.UserRepository {
	if s.userRepository == nil {
		s.userRepository = &UserRepository{
			Storage: s,
		}
	}
	return s.userRepository
}

func (s *Storage) Orders() storage.OrderRepository {
	if s.ordersRepository == nil {
		s.ordersRepository = &OrderRepository{
			Storage: s,
		}
	}
	return s.ordersRepository
}

//func (s *Storage) Withdrawals() storage.WithdrawRepository {
//	if s.withdrawRepository == nil {
//		s.withdrawRepository = &WithdrawRepository{
//			Storage: s,
//		}
//	}
//	return s.withdrawRepository
//}
