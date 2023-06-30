package http

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"

	"github.com/rubenvanstaden/nostr/core"
)

// A spoke is a valid user connection and therefore represents a user subscription.
type Spoke struct {

	// Wrapping the open websocket to read user events.
	conn *websocket.Conn

	// In-memory filters. Key is subscription Id.
	filters map[string]core.Filters

	// Stream to send broadcasted messages to user.
	send chan []byte

	// Store and filter events.
	repository core.Repository
}

// Write message to the spoke for the end user.
// This is done by the Relay when a message is place on it's broadcast channel.
func (s *Spoke) write(ctx context.Context) {

	defer s.conn.Close()

	for msg := range s.send {
		err := s.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

// Read messages coming from the spoke posted by the end user.
func (s *Spoke) read(ctx context.Context, relay *Relay) {

	defer s.conn.Close()

	for {

		// Get the raw message from the webconnect.
		_, raw, err := s.conn.ReadMessage()
		if err != nil {
			relay.unregister <- s
			return
		}

		// Decode the message type to match a decision pattern.
		msg := core.DecodeMessage(raw)

		switch msg.Type() {
		case "EVENT":
			var msg core.MessageEvent

			err = json.Unmarshal(raw, &msg)
			if err != nil {
				log.Fatalf("unable to unmarshal event: %v", err)
			}

			// We obvious;y want to see our own published event.
			s.filters = make(map[string]core.Filters)
			s.filters["0"] = core.Filters{
				core.Filter{Ids: []string{string(msg.Id)}},
			}

			// Persist the message for future subscribers.
			s.repository.Store(ctx, &msg.Event)

			relay.broadcast <- &msg.Event
		case "REQ":

			// 1. Parse the req message from the raw stream of data.
			var msg core.MessageReq
			err = json.Unmarshal(raw, &msg)
			if err != nil {
				log.Fatalf("unable to unmarshal event: %v", err)
			}

			if len(msg.Filters) == 0 {
				log.Println("no filters to be applied")
			}

			// 2. Query the event repository with the filter and get a set of events.
			for _, filter := range msg.Filters {

				// 				events, err := s.repository.FindByIdPrefix(ctx, filter.Ids)
				// 				if err != nil {
				// 					log.Fatalf("unable to retrieve events by IDs from repository: %v", err)
				// 				}

				events, err := s.repository.FindByAuthors(ctx, filter.Authors)
				if err != nil {
					log.Fatalf("unable to retrieve events by IDs from repository: %v", err)
				}

				if len(events) == 0 {
					log.Println("no events found")
				}

				// 3. Send these events to the current spoke's send channel.
				// There is no need to broadcast it to the hub, since we want to send the data to the current client.
				// We are basically just making a round trip to the event repository.
				for _, event := range events {

					msg := core.MessageEvent{
						SubscriptionId: msg.SubscriptionId,
						Event:          event,
					}

					bytes, err := json.Marshal(msg)
					if err != nil {
						log.Fatalln("unable to send REQ filtered messages")
					}

					s.send <- bytes
				}
			}

			// 4. Store the filter, over writting it if subId already exists.
			s.filters[msg.SubscriptionId] = msg.Filters
		}
	}
}
