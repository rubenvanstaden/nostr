package cli

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/env"
	"github.com/rubenvanstaden/nostr"
)

var (
	PRIVATE_KEY = env.String("NSEC")
)

type Connection struct {

	// Web socket connection between client and relay.
	socket *websocket.Conn

	// The connection owns the subscriptions.
	// Make a pointer, since we want to update the subscription event channel.
	subscriptions map[string]*Subscription

	// Write events from channel to connected relays.
	eventStream chan nostr.MessageEvent

	// Write request from channel to connected relay socket.
	reqStream chan nostr.MessageReq

	okStream chan nostr.MessageOk

	// Complete close connection
	done chan struct{}
}

func NewConnection(addr string) *Connection {
	// Connect to WebSocket server
	connection, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("websocket dial: ", err)
	}
	return &Connection{
		socket:        connection,
		subscriptions: make(map[string]*Subscription),
		eventStream:   make(chan nostr.MessageEvent),
		reqStream:     make(chan nostr.MessageReq),
		okStream:      make(chan nostr.MessageOk),
		done:          make(chan struct{}),
	}
}

// Listen to incoming events from remote relays by reading from socket.
// The caller should run this method in a goroutine.
func (s *Connection) Listen() error {

	// Listen to requests on the reqStream that should be broadcasted to relays.
	go func() {
		for {
			select {
			case <-s.done:
				return
			case event := <-s.eventStream:

				log.Println("Writing event to relays")

				// Marshal the signed event to a slice of bytes ready for transmission.
				bytes, err := json.Marshal(event)
				if err != nil {
					log.Fatalf("\nunable to marshal incoming EVENT: %#v", err)
				}

				// Transmit event message to the spoke that connects to the relays.
				err = s.socket.WriteMessage(websocket.TextMessage, bytes)
				if err != nil {
					// TODO: add an error channel pattern
					log.Fatalln(err)
				}

			case req := <-s.reqStream:

				log.Println("Writing request to relays")

				// Marshal to a slice of bytes ready for transmission.
				bytes, err := json.Marshal(req)
				if err != nil {
					log.Fatalf("\nunable to marshal incoming REQ: %#v", err)
				}

				// Transmit event message to the spoke that connects to the relays.
				err = s.socket.WriteMessage(websocket.TextMessage, bytes)
				if err != nil {
					// TODO: add an error channel pattern
					log.Fatalln(err)
				}

			}
		}
	}()

	// Read incoming messages on socket from relays.
	go func() {
		for {

			// Block for a status response from relays
			_, raw, err := s.socket.ReadMessage()
			if err != nil {
				log.Fatalln(err)
			}

			msg := nostr.DecodeMessage(raw)

			switch msg.Type() {
			case "EVENT":

				event := msg.(*nostr.MessageEvent)

				// Dispatch event to inmem subscription channel.
				if sub, ok := s.subscriptions[event.GetSubId()]; ok {
					sub.EventStream <- &event.Event
				}

				// Show relay response status after publishing an event.
			case "OK":
				ok := msg.(*nostr.MessageOk)
				s.okStream <- *ok

			// Close is end of new events.
			case "EOSE":
				continue
			}
		}
	}()

	log.Println("Connection established to relays")

	return nil
}

// Publish events to remote relays.
// Sign event before publishing
func (s *Connection) Publish(event nostr.Event) (*nostr.MessageOk, error) {

	// TODO: Maybe fix this

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

	go func() {
		s.eventStream <- nostr.MessageEvent{
            Event: event,
        }
	}()

	for {
		select {
		case <-s.done:
			return nil, nil
		case msg := <-s.okStream:
			if msg.GetEventId() == event.GetId() {
                log.Printf("OK: %v, eventId: %s, message: %s", msg.Ok, msg.GetEventId(), msg.Message)
				return &msg, nil
			}
		}
	}
}

func (s *Connection) Subscribe(filters nostr.Filters) (*Subscription, error) {

	// 1. Create a new subscription and take ownership

	sub := NewSubscription()

	s.subscriptions[sub.GetId()] = sub

	// 2. Fire a REQ to the relay.

	err := sub.Fire(filters, s.reqStream)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *Connection) Close() {
	// Disconnect from the WebSocket server
	err := s.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
	}
}
