package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vlad-marlo/gophermart/pkg/logger"

	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/poller"
	"github.com/vlad-marlo/gophermart/internal/server"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

const pollerQueueLimit = 20

func main() {
	// init logger
	log := logger.GetLogger()

	// init cfg
	cfg, err := config.New()
	if err != nil {
		log.Panicf("new config: %v", err)
	}

	ctx := context.Background()
	// init storage
	storage, err := sqlstore.New(ctx, log, cfg)
	if err != nil {
		log.Panicf("new sql store: %v", err)
	}
	p := poller.New(log, storage, pollerQueueLimit)
	go func() {
		log.Infof("starting server on %v", cfg.BindAddr)
		if err := server.Start(log, storage, cfg, p); err != nil {
			log.Panicf("start server: %v", err)
		}
	}()

	// creating interrupt chan for accepting os signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, os.Kill, syscall.SIGINT)

	// gracefully shut down
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		log.Info("got interrupt signal")
	case syscall.SIGTERM:
		log.Info("got terminate signal")
	case os.Kill:
		log.Info("got kill signal")
	case syscall.SIGINT:
		log.Infof("got int signal: %s", sig.String())
	default:
		log.Info("default")
	}

	p.Close()
	storage.Close()
	log.Info("server was closed successful")
}
