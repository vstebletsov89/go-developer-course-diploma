package storage

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage/repository"
)

type UserRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) *UserRepository {
	return &UserRepository{conn: conn}
}

func (r *UserRepository) RegisterUser(u *model.User) error {
	err := r.conn.QueryRow(
		"INSERT INTO users (login, password) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		u.Login,
		u.Password,
	).Scan(&u.ID)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		return repository.ErrorUserAlreadyExist
	}
	return nil
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
