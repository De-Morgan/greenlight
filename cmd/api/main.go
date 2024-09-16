package main

import (
	"expvar"
	"os"
	"runtime"
	"time"

	_ "github.com/lib/pq"
	"morgan.greenlight.nex/internal/logger"
)

var buildTime string

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

	expvar.NewString("version").Set(version)
	//Publish the number of active goroutine
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	//Publish the database connection pool statitics
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	//Public the current unix timestamp
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := newApplication(&cfg, db, logger)
	//Run DB migration
	runDBMigration(migrationUrl, cfg.db.dsn, logger)
	app.serve()
}
