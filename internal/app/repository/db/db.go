package repository

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"

	// Register postgres and file drivers
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Register pgx driver
	_ "github.com/jackc/pgx/stdlib"
)

func ConnectDB(DatabaseURI string) (*sql.DB, error) {
	db, err := sql.Open("pgx", DatabaseURI)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func DatabaseMigration(DatabaseURI string, path string) error {
	m, err := migrate.New(path, DatabaseURI)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
