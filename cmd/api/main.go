package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const (
	// Appication version number
	version      = "1.0.0"
	migrationUrl = "file://migration"
)

func main() {

	cfg := newConfig()
	db, err := openDB(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	app := newApplication(&cfg, db)
	//Run DB migration
	runDBMigration(migrationUrl, cfg.db.dsn)
	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	app.logger.Printf("starting %s server on %s", cfg.env, srv.Addr)

	err = srv.ListenAndServe()
	app.logger.Fatal(err)
}
