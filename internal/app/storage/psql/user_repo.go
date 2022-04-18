package psql

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage"
)

type UserRepository struct {
	Storage *Storage
}

func (r *UserRepository) RegisterUser(u *model.User) error {
	err := r.Storage.DB.QueryRow(
		"INSERT INTO users (login, password) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		u.Login,
		u.Password,
	).Scan(&u.ID)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		return storage.ErrorUserAlreadyExist
	}
	return nil
}

func (r *UserRepository) GetUser(login string) (*model.User, error) {
	u := &model.User{}
	err := r.Storage.DB.QueryRow(
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
		return nil, storage.ErrorUserNotFound
	}

	return u, nil
}
