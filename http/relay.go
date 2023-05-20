package http

import "log"

type Relay struct {
	spokes     map[*Spoke]bool
	broadcast  chan []byte
	register   chan *Spoke
	unregister chan *Spoke
}

func NewRelay() *Relay {
	return &Relay{
		spokes:     make(map[*Spoke]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Spoke),
		unregister: make(chan *Spoke),
	}
}

func (s *Relay) Run() {
	for {
		select {
		case client := <-s.register:
			s.spokes[client] = true
		case client := <-s.unregister:
			if _, ok := s.spokes[client]; ok {
				delete(s.spokes, client)
				close(client.send)
			}
		case message := <-s.broadcast:
			// Broadcast the message to all registered spokes.
			for client := range s.spokes {
				log.Println("Broadcast message to clients.")
				client.send <- message
			}
		}
	}
}
