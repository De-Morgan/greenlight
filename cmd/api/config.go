package main

import (
	"errors"
	"flag"
	"fmt"
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
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

func newConfig() config {
	var cfg config = config{
		env: development,
	}
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.Var(&cfg.env, "env", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "Postgres connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.gmail.com", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "<michaeladesola1410@gmail.com", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "password", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Morgan <michaeladesola1410@gmail.com>", "SMTP sender")

	showVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time: \t%s\n", buildTime)
		os.Exit(0)
	}

	return cfg
}
