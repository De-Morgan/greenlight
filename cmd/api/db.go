package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"morgan.greenlight.nex/internal/logger"
)

func openDB(cfg *config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runDBMigration(migrationUrl, dbSource string, logger *logger.Logger) {
	m, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		logger.PrintFatal("cannot create a new migration instance: ", map[string]any{
			"error": err,
		})
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.PrintFatal("failed to run migrate up: ", map[string]any{
			"error": err,
		})
	}
	logger.PrintInfo("db migrated successfully", nil)
}
