package main

import (
	"github.com/vlad-marlo/gophermart/internal/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/server"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

func main() {
	// init logger
	l := logger.GetLogger()
	l.Info("successfully init logger")

	// init cfg
	cfg, err := config.New()
	if err != nil {
		l.Fatalf("new config: %v", err)
	}

	store, err := sqlstore.New(l, cfg)
	if err != nil {
		l.Fatalf("new sql store: %v", err)
	}
	l.Info("successfully init sql storage")

	go func() {
		l.Infof("starting server on %v", cfg.BindAddr)
		if err := server.Start(l, store, cfg); err != nil {
			l.Fatalf("start server: %v", err)
		}
	}()

	// creating interrupt chan for accepting os signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// gracefully shut down
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		l.Info("got interrupt signal")
	case syscall.SIGTERM:
		l.Info("got terminate signal")
	}

	if err := store.Close(); err != nil {
		l.Fatalf("store: close: %v", err)
	}
	l.Info("storage was closed successful")
	l.Info("server was successful shut down")
}
