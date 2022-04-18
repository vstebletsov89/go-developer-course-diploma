package gophermart

import (
	"database/sql"
	"go-developer-course-diploma/internal/app/server"
	"go-developer-course-diploma/internal/app/storage/psql"
	"net/http"
)

const PostgreSQLUsersTable = `CREATE TABLE IF NOT EXISTS users (
    id bigserial NOT NULL PRIMARY KEY, 
    login text NOT NULL UNIQUE,
    password text NOT NULL
);`

func RunApp(cfg *Config) error {
	db, err := connectDB(cfg.DatabaseURI)
	if err != nil {
		return err
	}
	defer db.Close()

	storage := psql.NewStorage(db)
	srv := server.NewServer(storage, cfg.AccrualSystemAddress)

	return http.ListenAndServe(cfg.RunAddress, srv)
}

func connectDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	// test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	// migration users table
	_, err = db.Exec(PostgreSQLUsersTable)
	if err != nil {
		return nil, err
	}
	return db, nil
}
