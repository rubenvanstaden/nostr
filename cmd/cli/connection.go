package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr/core"
)

type Connection struct {
	// Web socket connection between client and relay.
	socket *websocket.Conn
	// Counter for subscriptions
	counter int
}

func NewConnection(addr string) *Connection {
	// Connect to WebSocket server
	connection, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("websocket dial: ", err)
	}
	return &Connection{
		socket:  connection,
		counter: 0,
	}
}

func (s *Connection) Publish(ev core.Event) (core.Status, error) {

	// TODO: Maybe fix this
	event := core.MessageEvent{
		Event: ev,
	}

	// Apply NIP-19 to decode user-friendly secrets.
	var sk string
	if _, s, e := crypto.DecodeBech32(PRIVATE_KEY); e == nil {
		sk = s.(string)
	}
	if pub, e := crypto.GetPublicKey(sk); e == nil {
		// Set public with which the event wat pushed.
		event.PubKey = pub
	}
	// We have to sign last, since the signature is dependent on the event content.
	event.Sign(sk)

	// Marshal the signed event to a slice of bytes ready for transmission.
	msg, err := json.Marshal(event)
	if err != nil {
		log.Fatalln("unable to marchal incoming event")
	}

	// Transmit event message to the spoke that connects to the relays.
	err = s.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return core.StatusFail, err
	}

	// Streaming reponses from the connected relay.
	// Wait till be get the OK from all relays.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, raw, err := s.socket.ReadMessage()
			if err != nil {
				log.Fatalln(err)
				return
			}
			msg := core.DecodeMessage(raw)
			switch msg.Type() {
			case "OK":
				e := msg.(*core.MessageResult)
				log.Printf("[\033[32m*\033[0m] Relay")
				log.Printf("  status: OK")
				log.Printf("  message: %s", e.Message)
				return
			default:
				log.Fatalln("unknown message type from RELAY")
				return
			}
		}
	}()
	wg.Wait()

	return core.StatusOK, nil
}

func (s *Connection) Request(ev core.Event) (core.Status, error) {

	return core.StatusOK, nil
}

func (s *Connection) Close() {
	// Disconnect from the WebSocket server
	err := s.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
	}
}
