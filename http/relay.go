package http

import (
	"encoding/json"
	"log"
	"noztr/core"
)

type Relay struct {
	spokes     map[*Spoke]bool
	broadcast  chan *core.Event
	register   chan *Spoke
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
					if filters.Match(event) {
						log.Println("Broadcast message to clients.")

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
