package main

import (
	"context"
	"github.com/vlad-marlo/gophermart/internal/store"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/poller"
	"github.com/vlad-marlo/gophermart/internal/server"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

const pollerQueueLimit = 30

func main() {
	ctx := context.Background()

	// init logger
	log := logger.GetLogger()

	// init cfg
	cfg, err := config.New()
	if err != nil {
		log.Panicf("new config: %v", err)
	}

	// init storage
	storage, err := sqlstore.New(ctx, log, cfg)
	if err != nil {
		log.Panicf("new sql store: %v", err)
	}

	p := poller.New(log, storage, cfg, pollerQueueLimit)
	s := server.New(log, storage, cfg, p)

	defer tearDown(log, storage, p)

	go func() {
		if err := http.ListenAndServe(cfg.BindAddr, s.Router); err != nil {
			log.Panicf("start server: %v", err)
		}
	}()
}

func tearDown(logger logger.Logger, store store.Storage, p *poller.OrderPoller) {
	// creating interrupt chan for accepting os signals for graceful shut down
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)

	sig := <-interrupt

	p.Close()
	store.Close()

	logger.WithField("signal", sig.String()).Info("graceful shut down")
}
