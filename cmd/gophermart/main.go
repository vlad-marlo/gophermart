package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/vlad-marlo/gophermart/internal/pkg/logger"

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
		l.Panicf("new config: %v", err)
	}

	// init store
	store, err := sqlstore.New(l, cfg)
	if err != nil {
		l.Panicf("new sql store: %v", err)
	}
	l.Info("successfully init sql storage")

	go func() {
		l.Infof("starting server on %v", cfg.BindAddr)
		if err := server.Start(l, store, cfg); err != nil {
			l.Panicf("start server: %v", err)
		}
	}()

	// creating interrupt chan for accepting os signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, os.Kill, syscall.SIGSTOP)

	// gracefully shut down
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		l.Info("got interrupt signal")
	case syscall.SIGTERM:
		l.Info("got terminate signal")
	case os.Kill:
		l.Info("got kill signal")
	default:
		l.Info("default")
	}

	if err := store.Close(); err != nil {
		l.Panicf("store: close: %v", err)
	}
	l.Info("storage was closed successful")
	l.Info("server was successful shut down")
}
