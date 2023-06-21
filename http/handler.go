package http

import (
	"context"
	"log"
	"net/http"
	"noztr/core"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	// Upgrade the http protocol to a websocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	spoke := &Spoke{
		conn:       conn,
		filters:    make(map[string]core.Filters),
		send:       make(chan []byte),
		repository: s.repository,
	}

	// Register the client to the relay.
	s.relay.register <- spoke

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		spoke.write(ctx)
	}()
	go func() {
		defer wg.Done()
		spoke.read(ctx, s.relay)
	}()
	wg.Wait()
}
