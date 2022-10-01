package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vlad-marlo/gophermart/internal/config"
)

func main() {
	config, err := config.New()
	if err != nil {
	}
	// creating interrupt chan for accepting os signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// gracefull shut down
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		log.Print("got interrupt signal")
	case syscall.SIGTERM:
		log.Print("got terminate signal")
	}

}
