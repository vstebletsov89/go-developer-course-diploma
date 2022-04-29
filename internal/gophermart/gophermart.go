package gophermart

import (
	"context"
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"go-developer-course-diploma/internal/accrual"
	"go-developer-course-diploma/internal/configs"
	"go-developer-course-diploma/internal/controller"
	"go-developer-course-diploma/internal/server"
	"go-developer-course-diploma/internal/service/auth"
	storage2 "go-developer-course-diploma/internal/storage"
	"net/http"
	"time"
)

//go:embed migrations/*.sql
var migrationsContent embed.FS

func RunApp(cfg *configs.Config) error {
	// init global logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return err
	}
	logger.SetLevel(level)

	// debug config
	logger.Debugf("%+v\n\n", cfg)

	db, err := connectDB(cfg.DatabaseURI, logger)
	if err != nil {
		logger.Infof("connectDB error: %s", err)
		return err
	}
	defer db.Close()

	// run db migration
	if err := RunMigrations(db, migrationsContent, logger); err != nil {
		logger.Infof("Migration error: %s", err)
		return err
	}

	userStore := storage2.NewUserRepository(db)
	orderStore := storage2.NewOrderRepository(db)
	transactionStore := storage2.NewTransactionRepository(db)
	userAuthStore := auth.NewUserAuthorizationStore()
	c := controller.NewController(cfg, logger, userStore, orderStore, transactionStore, userAuthStore)

	// create accrual provider
	p := accrual.NewAccrualClient(cfg, logger, orderStore, transactionStore)

	// check pending orders
	go p.CheckPendingOrders(context.Background())

	srv := server.NewServer(c)
	return http.ListenAndServe(cfg.RunAddress, srv)
}

func RunMigrations(db *sql.DB, migrationsContent embed.FS, logger *logrus.Logger) error {
	logger.Infof("Start db migration")
	goose.SetBaseFS(migrationsContent)
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	logger.Infof("Migration completed")
	return nil
}

func connectDB(databaseURL string, logger *logrus.Logger) (*sql.DB, error) {
	logger.Infof("Open DB connection")
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}
