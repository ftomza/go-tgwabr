package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tgwabr"
)

func main() {

	shutDownHandler := tgwabr.Init()

	log.Println("Service is UP")

	defer func() {
		log.Println("Shutdown instances")
		shutDownHandler()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
