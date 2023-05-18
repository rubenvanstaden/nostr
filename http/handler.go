package http

import (
	//"context"
	//"encoding/json"
	//"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	//"noztr/core"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {

	//ctx := context.Background()

	// Upgrade the http protocol to a websocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte),
	}

	s.hub.register <- client

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		client.write()
	}()
	go func() {
		defer wg.Done()
		client.read(s.hub)
	}()
	wg.Wait()
}
