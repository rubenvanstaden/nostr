package cli

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"

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

func (s *Connection) Publish(ev nostr.Event) (nostr.Status, error) {

	// TODO: Maybe fix this
	event := nostr.MessageEvent{
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
		return nostr.StatusFail, err
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
			msg := nostr.DecodeMessage(raw)
			switch msg.Type() {
			case "OK":
				e := msg.(*nostr.MessageResult)
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

	return nostr.StatusOK, nil
}

// TODO: Currently only returing a single events. Should be a stream
func (s *Connection) Request(ctx context.Context, filters nostr.Filters) (*nostr.Event, error) {

	var req nostr.MessageReq
	req.SubscriptionId = "follow" + ":" + strconv.Itoa(s.counter)
	req.Filters = filters

	// Marshal to a slice of bytes ready for transmission.
	msg, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("\nunable to marshal incoming REQ event: %#v", err)
	}

	// Transmit event message to the spoke that connects to the relays.
	err = s.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return nil, err
	}

    // Stream messages from websocket to inmem channel until context done.
    orDone := func() <-chan *nostr.Event {
        valStream := make(chan *nostr.Event)
        go func() {
            defer close(valStream)
            for {
                select {
                    case <-ctx.Done():
                        return
                    default:
                        valStream <- read(s.socket)
                }
            }
        }()
        return valStream
    }

    for val := range orDone() {
        log.Println("VAL")
        log.Println(val)
    }

	return nil, nil
}

func (s *Connection) Close() {
	// Disconnect from the WebSocket server
	err := s.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
	}
}

func read(connection *websocket.Conn) *nostr.Event {

	// Block for a status response from relays
	_, raw, err := connection.ReadMessage()
	if err != nil {
		log.Fatalln(err)
	}

	m := nostr.DecodeMessage(raw)

	switch m.Type() {
	case "EVENT":

		event := m.(*nostr.MessageEvent)

		switch event.Kind {
		case nostr.KindTextNote:
			return &event.Event
		case nostr.KindSetMetadata:
			_, err := nostr.ParseMetadata(event.Event)
			if err != nil {
				log.Fatalf("unable to pull profile: %#v", err)
			}
			return &event.Event
		}

	default:
		log.Fatalln("unknown message type from RELAY")
	}
    return nil
}

