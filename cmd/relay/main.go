package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/rubenvanstaden/env"

	"github.com/rubenvanstaden/nostr/mongodb"
	"github.com/rubenvanstaden/nostr/relay"
)

var (
	RELAY_URL      = env.String("RELAY_URL")
	REPOSITORY_URL = env.MongoAddr("REPOSITORY_URL")
)

func main() {

	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	repository := mongodb.New(REPOSITORY_URL, "noztr", "events")

	hub := relay.NewHub()
	go hub.Run()

	s := relay.NewServer(RELAY_URL, hub, repository)

	// Start the HTTP server.
	err := s.Open()
	if err != nil {
		log.Fatalf("unable to open connection to relay: %v\n", err)
	}

	log.Printf("Serving on address: %s\n", RELAY_URL)

	// Wait for CTRL-C.
	<-ctx.Done()

	// Shutdown HTTP server
	err = s.Close()
	if err != nil {
		log.Fatalf("unable to CLOSE connection to relay: %v\n", err)
	}

	log.Printf("Shutdown complete: %s\n", RELAY_URL)
}
