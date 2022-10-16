package main

import (
	"context"
	"github.com/vlad-marlo/gophermart/pkg/logger"
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

	// init cfg
	cfg, err := config.New()
	if err != nil {
		l.Panicf("new config: %v", err)
	}

	ctx := context.Background()
	// init store
	store, err := sqlstore.New(ctx, l, cfg)
	if err != nil {
		l.Panicf("new sql store: %v", err)
	}

	go func() {
		l.Infof("starting server on %v", cfg.BindAddr)
		if err := server.Start(l, store, cfg); err != nil {
			l.Panicf("start server: %v", err)
		}
	}()

	// creating interrupt chan for accepting os signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, os.Kill, syscall.SIGINT)

	// gracefully shut down
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		l.Info("got interrupt signal")
	case syscall.SIGTERM:
		l.Info("got terminate signal")
	case os.Kill:
		l.Info("got kill signal")
	case syscall.SIGINT:
		l.Infof("got int signal: %s", sig.String())
	default:
		l.Info("default")
	}

	store.Close()
	l.Info("server was closed successful")
}
