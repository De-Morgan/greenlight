package main

import (
	"errors"
	"flag"
	"os"
	"slices"
	"time"
)

type workingEnv string

func (w *workingEnv) String() string { return string(*w) }
func (w *workingEnv) Set(s string) error {
	if !slices.Contains([]workingEnv{
		development, staging, production,
	}, workingEnv(s)) {
		return errors.New("invalid working environment")
	}
	*w = workingEnv(s)
	return nil
}

const (
	development workingEnv = "development"
	staging     workingEnv = "staging"
	production  workingEnv = "production"
)

// Application configuration
type config struct {
	port int
	env  workingEnv
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

func newConfig() config {
	var cfg config = config{
		env: development,
	}
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.Var(&cfg.env, "env", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("GREENLIGHT_DB_DSN"), "Postgres connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")
	flag.Parse()

	return cfg
}
