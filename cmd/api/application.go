package main

import (
	"database/sql"

	"morgan.greenlight.nex/internal/data"
	"morgan.greenlight.nex/internal/logger"
)

// Define an application struct to hold the dependencies for
// HTTP handlers, helpers and middlewares
type application struct {
	config config
	logger *logger.Logger
	models data.Models
}

func newApplication(cfg *config, db *sql.DB, logger *logger.Logger) *application {
	return &application{
		config: *cfg,
		logger: logger,
		models: data.NewModels(db),
	}

}
