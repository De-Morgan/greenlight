package main

import (
	"os"

	_ "github.com/lib/pq"
	"morgan.greenlight.nex/internal/logger"
)

const (
	// Appication version number
	version      = "1.0.0"
	migrationUrl = "file://migration"
)

func main() {
	logger := logger.New(os.Stdout, logger.LevelInfo)
	cfg := newConfig()
	db, err := openDB(&cfg)
	if err != nil {
		logger.PrintFatal(err.Error(), nil)
	}
	defer db.Close()
	app := newApplication(&cfg, db, logger)
	//Run DB migration
	runDBMigration(migrationUrl, cfg.db.dsn, logger)
	app.serve()
}
