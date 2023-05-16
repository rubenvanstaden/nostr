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

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/echo"}

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

	// Send messages to the WebSocket server
	for i := 0; i < 5; i++ {
		message := []byte("Hello, server!")
		err = conn.WriteMessage(websocket.TextMessage, message)
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
