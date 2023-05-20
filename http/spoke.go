package http

import (
	//"encoding/json"
	"encoding/json"
	"fmt"
	"log"
	"noztr/core"

	//"log"
	//"noztr/core"

	"github.com/gorilla/websocket"
)

// A spoke is a valid user connection. And therefore can represent a subscription.
type Spoke struct {

	// Wrapping the open websocket to read user events.
	conn *websocket.Conn

	// Inmem filters. Key is subscription Id.
	filters map[string][]core.Filter

	// Channel to send broadcasted messages to user.
	send chan []byte
}

// Write message to the spoke for the end user. This is done by the Relay when a message is place on it's broadcast channel.
func (s *Spoke) write() {

	defer s.conn.Close()

	for {
		select {
		case msg := <-s.send:
			resp := fmt.Sprintf("{msg from user A: %s}", msg)
			err := s.conn.WriteMessage(websocket.TextMessage, []byte(resp))
			if err != nil {
				return
			}
		}
	}
}

// Read messages coming from the spoke posted by the end user.
func (s *Spoke) read(relay *Relay) {

	defer s.conn.Close()

	for {
		_, raw, err := s.conn.ReadMessage()
		if err != nil {
			relay.unregister <- s
			return
		}

		fmt.Printf("Msg read: %s\n", raw)

		msg := core.DecodeMessage(raw)

		switch msg.Type() {
		case "EVENT":
			var msg core.MessageEvent

			err = json.Unmarshal(raw, &msg)
			if err != nil {
				log.Fatalf("unable to unmarshal event: %v", err)
			}

			fmt.Printf("Event parsed: %#v\n", msg)
		case "REQ":
		}

		relay.broadcast <- raw
		//hub.broadcast <- msg
	}
}
