package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := log.New(os.Stderr, "Bee: ", log.LstdFlags)
	logger.Println("Starting...")
	NewTiZigbee("a1b2c3", logger)
	logger.Println("Started...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logger.Println("Ending...")
}
