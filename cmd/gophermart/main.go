package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/server"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

func main() {
	// init logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)
	logger.Debug("sucessfully init logger")

	// init config
	config, err := config.New()
	if err != nil {
		logger.Fatalf("new config: %v", err)
	}
	store, err := sqlstore.New(logger, config)
	if err != nil {
		logger.Fatalf("new sql store: %v", err)
	}
	logger.Debug("sucessfully init sql storage")
	go func() {
		logger.Debugf("starting server on %v", config.BindAddr)
		if err := server.Start(logger, store, config); err != nil {
			logger.Fatalf("start server: %v", err)
		}
	}()
	// creating interrupt chan for accepting os signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// gracefull shut down
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		logger.Info("got interrupt signal")
	case syscall.SIGTERM:
		logger.Info("got terminate signal")
	}

	if err := store.Close(); err != nil {
		logger.Fatalf("store: close: %v", err)
	}
}
