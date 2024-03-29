package relay

import (
	"encoding/json"
	"log"

	"github.com/rubenvanstaden/nostr"
)

type Hub struct {

	// Data structure to in-memory store registered spokes.
	spokes map[*Spoke]bool

	// Stream to broadcast event message to registered clients.
	broadcast chan *nostr.Event

	// Stream to concurrently handle client spoke registration.
	register chan *Spoke

	// Stream to concurrently unregister a client spoke.
	unregister chan *Spoke
}

func NewHub() *Hub {
	return &Hub{
		spokes:     make(map[*Spoke]bool),
		broadcast:  make(chan *nostr.Event),
		register:   make(chan *Spoke),
		unregister: make(chan *Spoke),
	}
}

// TODO: Add context with done for shutdown.
func (s *Hub) Run() {
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

					if filters.Match(event) {

						msg := nostr.MessageEvent{
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
