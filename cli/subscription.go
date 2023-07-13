package cli

import (
	"log"
	"strconv"
	"sync/atomic"

	"github.com/rubenvanstaden/nostr"
)

var subId atomic.Int32

type Subscription struct {

	// Keep track of total subscriptions
	counter int

	// Place events from read socket onto inmem channel.
	EventStream chan *nostr.Event

	// Close subscription when EOSE event read from socket.
	done chan struct{}
}

func NewSubscription() *Subscription {

	// Increment the subscription counter.
	counter := subId.Add(1)

	return &Subscription{
		counter:     int(counter),
		EventStream: make(chan *nostr.Event),
		done:        make(chan struct{}),
	}
}

func (s *Subscription) GetId() string {
	return strconv.Itoa(s.counter)
}

// Sends a REQ to the relay via the connection.
func (s *Subscription) Fire(filters nostr.Filters, reqStream chan<- nostr.MessageReq) error {

	log.Println("Fire subscription request to relays")

	var req nostr.MessageReq
	req.SubscriptionId = s.GetId()
	req.Filters = filters

	select {
	case reqStream <- req:
	case <-s.done:
		return nil
	}

	return nil
}

func (s *Subscription) Close() error {
	return nil
}
