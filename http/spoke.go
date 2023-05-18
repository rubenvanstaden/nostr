package http

import (
	"encoding/json"
	"fmt"
	"noztr/core"

	"github.com/gorilla/websocket"
)

type Client struct {

    // Wrapping the open websocket to read user events.
	conn *websocket.Conn

    // Channel to send broadcasted messages to user.
	send chan []byte
}

// Write message to the spoke for the end user. This is done by the Hub when a message is place on it's broadcast channel.
func (s *Client) write() {

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
func (s *Client) read(hub *Hub) {

	defer s.conn.Close()

	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			hub.unregister <- s
			return
		}

 		var event core.Event
 		err = json.Unmarshal(msg, &event)
 		if err != nil {
 			panic(err)
 		}
 
 		fmt.Printf("Event parsed: %#v\n", event)

		hub.broadcast <- msg
	}
}
