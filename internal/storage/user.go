package storage

import (
	"database/sql"
	"go-developer-course-diploma/internal/model"
	"go-developer-course-diploma/internal/storage/repository"
)

type UserRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) *UserRepository {
	return &UserRepository{conn: conn}
}

func (r *UserRepository) RegisterUser(u *model.User) (int64, error) {
	err := r.conn.QueryRow(
		"INSERT INTO users (login, password) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		u.Login,
		u.Password,
	).Scan(&u.ID)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		return 0, repository.ErrorUserAlreadyExist
	}
	return u.ID, nil
}

func (r *UserRepository) GetUser(login string) (*model.User, error) {
	u := &model.User{}
	err := r.conn.QueryRow(
		"SELECT id, login, password FROM users WHERE login = $1",
		login,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, repository.ErrorUserNotFound
	}

	return u, nil
}
