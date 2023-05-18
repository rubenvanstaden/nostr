package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/rubenvanstaden/env"

	"noztr/http"
)

var (
	RELAY_URL      = env.WebsocketAddr("RELAY_URL")
	REPOSITORY_URL = env.MongoAddr("REPOSITORY_URL")
)

func main() {

	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	hub := http.NewHub()
	go hub.Run()

	s := http.NewServer(RELAY_URL, hub)

	// Start the HTTP server.
	err := s.Open()
	if err != nil {
		panic(err)
	}

	log.Printf("Serving on address: %s\n", RELAY_URL)

	// Wait for CTRL-C.
	<-ctx.Done()

	// Shutdown HTTP server
	err = s.Close()
	if err != nil {
		panic(err)
	}

	log.Printf("Shutdown complete: %s\n", RELAY_URL)
}
