package main

import (
	"database/sql"
	"sync"

	"morgan.greenlight.nex/internal/data"
	"morgan.greenlight.nex/internal/logger"
	"morgan.greenlight.nex/internal/mailer"
)

// Define an application struct to hold the dependencies for
// HTTP handlers, helpers and middlewares
type application struct {
	config config
	logger *logger.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func newApplication(cfg *config, db *sql.DB, logger *logger.Logger) *application {
	return &application{
		config: *cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

}
