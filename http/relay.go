package http

import (
	"encoding/json"
	"log"
	"noztr/core"
)

type Relay struct {

	// Data structure to in-memory store registered spokes.
	spokes map[*Spoke]bool

	// Stream to broadcast event message to registered clients.
	broadcast chan *core.Event

	// Stream to concurrently handle client spoke registration.
	register chan *Spoke

	// Stream to concurrently unregister a client spoke.
	unregister chan *Spoke
}

func NewRelay() *Relay {
	return &Relay{
		spokes:     make(map[*Spoke]bool),
		broadcast:  make(chan *core.Event),
		register:   make(chan *Spoke),
		unregister: make(chan *Spoke),
	}
}

// TODO: Add context with done for shutdown.
func (s *Relay) Run() {
	for {
		select {
		case spoke := <-s.register:
			s.spokes[spoke] = true
		case client := <-s.unregister:
			if _, ok := s.spokes[client]; ok {
				delete(s.spokes, client)
				close(client.send)
			}
		case event := <-s.broadcast:

			// Broadcast the message to all registered spokes.
			for spoke := range s.spokes {

				// Check is message passes the spokes filters.
				for subId, filters := range spoke.filters {

					log.Println(subId)

					if filters.Match(event) {

						log.Println("aweeeee")

						msg := core.MessageEvent{
							SubscriptionId: subId,
							Event:          *event,
						}

						bytes, err := json.Marshal(msg)
						if err != nil {
							log.Fatalln("unable to broadcast filtered message")
						}

						spoke.send <- bytes
					}
				}
			}
		}
	}
}
