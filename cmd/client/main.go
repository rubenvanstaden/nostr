package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

func main() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Failed to establish WebSocket connection:", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	// Read messages from the WebSocket server
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message from WebSocket server:", err)
				return
			}
			log.Printf("Received from Relay: %s\n", message)
		}
	}()

	e1 := `{"id":"1","kind":"1","content":"hello world 1"}`
	e2 := `{"id":"2","kind":"1","content":"hello world 2"}`

	// Send messages to the WebSocket server
	for _, msg := range []string{e1, e2} {
		err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("Failed to send message to WebSocket server:", err)
			return
		}
	}

	select {
	case <-interrupt:
		log.Println("Interrupt signal received, closing WebSocket connection...")
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Failed to send close message to WebSocket server:", err)
		}
		<-done
		return
	}
}
