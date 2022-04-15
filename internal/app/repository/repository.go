package repository

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

func DatabaseMigration(databaseURI string, logger *logrus.Logger) error {
	logger.Info("Start database migration")
	m, err := migrate.New(
		"file://internal/app/repository/migrations",
		databaseURI)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		return err
	}
	logger.Info("Database migration completed")
	return nil
}
