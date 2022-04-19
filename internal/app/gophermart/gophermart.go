package gophermart

import (
	"context"
	"database/sql"
	"github.com/sirupsen/logrus"
	"go-developer-course-diploma/internal/app/accrual"
	"go-developer-course-diploma/internal/app/configs"
	"go-developer-course-diploma/internal/app/controller"
	"go-developer-course-diploma/internal/app/server"
	"go-developer-course-diploma/internal/app/storage/psql"
	"log"
	"net/http"
	"time"
)

const PostgreSQLUsersTable = `CREATE TABLE IF NOT EXISTS users (
    id bigserial NOT NULL PRIMARY KEY, 
    login text NOT NULL UNIQUE,
    password text NOT NULL
);`

const PostgreSQLOrdersTable = `CREATE TABLE IF NOT EXISTS orders (
    id bigserial NOT NULL PRIMARY KEY, 
    login text NOT NULL,
    number text NOT NULL UNIQUE,
    status text NOT NULL,
    accrual numeric DEFAULT 0,
    uploaded_at timestamptz NOT NULL
);`

func RunApp(cfg *configs.Config) error {
	// init global logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal("Failed to parse log level")
	}
	logger.SetLevel(level)

	db, err := connectDB(cfg.DatabaseURI, logger)
	if err != nil {
		return err
	}
	defer db.Close()

	s := psql.NewStorage(db)
	c := controller.NewController(cfg, s, logger)

	// check pending orders
	go accrual.UpdatePendingOrders(c, context.Background())

	srv := server.NewServer(c)
	return http.ListenAndServe(cfg.RunAddress, srv)
}

func connectDB(databaseURL string, logger *logrus.Logger) (*sql.DB, error) {
	logger.Infof("Open DB connection")
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// test connection
	logger.Infof("Ping connection")
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// migration users table
	logger.Infof("Database migration")
	_, err = db.Exec(PostgreSQLUsersTable)
	if err != nil {
		return nil, err
	}
	// migration orders table
	_, err = db.Exec(PostgreSQLOrdersTable)
	if err != nil {
		return nil, err
	}

	logger.Infof("Database migration completed")
	return db, nil
}
