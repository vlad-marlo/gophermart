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

const pollerQueueLimit = 10

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

	p := poller.New(log, storage, cfg, pollerQueueLimit)

	go func() {
		if err := server.Start(log, storage, cfg, p); err != nil {
			log.Panicf("start server: %v", err)
		}
	}()

	// creating interrupt chan for accepting os signals for graceful shut down
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)

	sig := <-interrupt

	p.Close()
	storage.Close()

	log.WithField("signal", sig.String()).Info("graceful shut down")
}
