package psql

import (
	"database/sql"
	_ "github.com/lib/pq"
	"go-developer-course-diploma/internal/app/storage"
)

type Storage struct {
	DB             *sql.DB
	userRepository *UserRepository
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		DB: db,
	}
}

func (s *Storage) User() storage.UserRepository {
	if s.userRepository == nil {
		s.userRepository = &UserRepository{
			Storage: s,
		}
	}

	return s.userRepository
}
