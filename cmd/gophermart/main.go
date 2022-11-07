package main

import (
	"context"
	"github.com/vlad-marlo/gophermart/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vlad-marlo/gophermart/internal/config"
	"github.com/vlad-marlo/gophermart/internal/poller"
	"github.com/vlad-marlo/gophermart/internal/server"
	"github.com/vlad-marlo/gophermart/internal/store/sqlstore"
)

const pollInterval = 1 * time.Second

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
	defer storage.Close()

	p := poller.New(log, storage, cfg, pollInterval)
	defer p.Close()
	s := server.New(log, storage, cfg)

	go func() {
		if err := http.ListenAndServe(cfg.BindAddr, s.Router); err != nil {
			log.Panicf("start server: %v", err)
		}
	}()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)

	sig := <-interrupt

	log.WithField("signal", sig.String()).Info("graceful shut down")
}
