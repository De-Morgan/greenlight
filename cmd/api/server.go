package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() {
	showdownErr := make(chan error)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(app.logger, "", 0),
	}

	go app.catchSignal(srv, showdownErr)

	app.logger.PrintInfo("starting server", map[string]any{
		"env": app.config.env, "addr": srv.Addr,
	})

	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		app.logger.PrintFatal(err.Error(), nil)
	}
	err = <-showdownErr
	if err != nil {
		app.logger.PrintFatal("problem with graceful shutdown", map[string]any{
			"error": err.Error(),
		})
	} else {
		app.logger.PrintInfo("stopped server", map[string]any{"addr": srv.Addr})
	}
}

func (app *application) catchSignal(srv *http.Server, showdownErr chan<- error) {
	quit := make(chan os.Signal, 1)

	// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and
	// relay them to the quit channel. Any other signals will not be caught by
	// signal.Notify() and will retain their default behavior.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	s := <-quit
	app.logger.PrintInfo("caught signal", map[string]any{"signal": s.String()})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	showdownErr <- srv.Shutdown(ctx)

}
