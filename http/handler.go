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

	//	for {
	//		// Read message from the WebSocket client
	//		messageType, msg, err := conn.ReadMessage()
	//		if err != nil {
	//			log.Println(err)
	//			break
	//		}
	//
	//		fmt.Printf("Received from Client: %s\n", msg)
	//
	//		var event core.Event
	//		err = json.Unmarshal(msg, &event)
	//		if err != nil {
	//			panic(err)
	//		}
	//
	//		fmt.Printf("Event parsed: %#v\n", event)
	//
	//		//repository := mongodb.New(REPOSITORY_URL, "noztr", "nevents")
	//		//repository.Store(ctx, &event)
	//
	//		// Echo the message back to the client
	//		resp := fmt.Sprintf("Event published: {id: %s}", event.Id)
	//		err = conn.WriteMessage(messageType, []byte(resp))
	//		if err != nil {
	//			log.Println(err)
	//			break
	//		}
	//	}
}
