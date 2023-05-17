package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/env"

	"noztr/core"
	"noztr/mongodb"
)

var (
	REPOSITORY_URL = env.String("REPOSITORY_URL")
)

// Upgrade the HTTP connection to a WebSocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	// Upgrade the connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		// Read message from the WebSocket client
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Print the received message
		fmt.Printf("Received from Client: %s\n", msg)

		var event core.Event
		err = json.Unmarshal(msg, &event)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Event parsed: %#v\n", event)

		repository := mongodb.New(REPOSITORY_URL, "noztr", "nevents")
		repository.Store(ctx, &event)

		// Echo the message back to the client
        resp := fmt.Sprintf("Event published: {id: %s}", event.Id)
		err = conn.WriteMessage(messageType, []byte(resp))
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func main() {

	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
