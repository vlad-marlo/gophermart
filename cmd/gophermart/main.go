package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	switch sig {
	case os.Interrupt:
		log.Print("got interrupt signal")
	case syscall.SIGTERM:
		log.Print("got terminate signal")
	}

}
