package main

import (
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/hooks/auth"
	"github.com/mochi-co/mqtt/v2/listeners"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := log.New(os.Stderr, "Bus: ", log.LstdFlags)
	logger.Println("Starting...")
	NewBus()
	logger.Println("Started...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logger.Println("Ending...")
}

func NewBus() {
	server := *mqtt.New(nil)
	server.AddHook(new(auth.AllowHook), nil)
	tcpSocket := listeners.NewTCP("tcp", ":1883", &listeners.Config{})
	server.AddListener(tcpSocket)
	server.Serve()
}
