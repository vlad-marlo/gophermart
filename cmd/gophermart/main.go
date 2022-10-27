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
	p := poller.New(log, storage, cfg, pollerQueueLimit)
	go func() {
		log.Infof("starting server on %v", cfg.BindAddr)
		if err := server.Start(log, storage, cfg, p); err != nil {
			log.Panicf("start server: %v", err)
		}
	}()

	// creating interrupt chan for accepting os signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, os.Kill, syscall.SIGINT, syscall.SIGSEGV)

	// gracefully shut down
	var stringSignal string
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		stringSignal = "interrupt"
	case syscall.SIGTERM:
		stringSignal = "terminate"
	case os.Kill:
		stringSignal = "kill"
	case syscall.SIGINT:
		stringSignal = "int"
	case syscall.SIGSEGV:
		stringSignal = "segmentation violation"
	default:
		stringSignal = "unknown"
	}

	p.Close()
	storage.Close()
	log.WithField("signal", stringSignal).Info("graceful shut down")
}
