package main

import (
	"encoding/json"
	"flag"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr/core"
)

func NewFollow(cc *Connection) *Follow {

	gc := &Follow{
		fs: flag.NewFlagSet("follow", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.StringVar(&gc.ls, "ls", "", "event text note of Kind 1")
	gc.fs.StringVar(&gc.add, "add", "", "event text note of Kind 0")
	gc.fs.StringVar(&gc.remove, "remove", "", "event text note of Kind 2")

	return gc
}

type Follow struct {
	fs *flag.FlagSet
	cc *Connection

	ls     string
	add    string
	remove string
}

func (g *Follow) Name() string {
	return g.fs.Name()
}

func (g *Follow) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Follow) Run() error {

	if s.ls != "" {
		log.Println("[ls] not implemented")
	}

	if s.add != "" {
		s.subscribe(s.add)
	}

	if s.remove != "" {
		log.Println("[remove] not implemented")
	}

	return nil
}

func (s *Follow) subscribe(npub string) error {

	// Decode npub using NIP-19
	_, pk, err := crypto.DecodeBech32(npub)
	if err != nil {
		log.Fatalf("\nunable to decode npub: %#v", err)
	}

	f := core.Filter{
		Authors: []string{pk.(string)},
	}

	var req core.MessageReq
	req.SubscriptionId = "follow" + ":" + strconv.Itoa(s.cc.counter)
	req.Filters = core.Filters{f}

	// Marshal to a slice of bytes ready for transmission.
	msg, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("\nunable to marshal incoming REQ event: %#v", err)
	}

	log.Printf("[\033[32m*\033[0m] Client - Request to follow (npub: %s...)", npub[:10])

	// Transmit event message to the spoke that connects to the relays.
	err = s.cc.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}
