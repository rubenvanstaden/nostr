package main

import (
	"encoding/json"
	"flag"
	"log"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr/core"
)

func NewFollow(cc *Connection) *Follow {

	gc := &Follow{
		fs: flag.NewFlagSet("follow", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.StringVar(&gc.ls, "ls", "", "list all following users")
	gc.fs.StringVar(&gc.add, "add", "", "user public key to add to following list")
	gc.fs.StringVar(&gc.remove, "remove", "", "remove user via public key")

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
		Kinds:   []uint32{core.KindTextNote},
		Limit:   5,
	}

	var req core.MessageReq
	req.SubscriptionId = "follow" + ":" + strconv.Itoa(s.cc.counter)
	req.Filters = core.Filters{f}

	// Marshal to a slice of bytes ready for transmission.
	msg, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("\nunable to marshal incoming REQ event: %#v", err)
	}

	log.Printf("[\033[32m*\033[0m] Client")
	log.Printf("  request to follow (npub: %s...)", npub[:10])

	// Transmit event message to the spoke that connects to the relays.
	err = s.cc.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	// Streaming reponses from the connected relay.

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, raw, err := s.cc.socket.ReadMessage()
			if err != nil {
				log.Fatalln(err)
				return
			}
			msg := core.DecodeMessage(raw)
			switch msg.Type() {
			case "EVENT":
				event := msg.(*core.MessageEvent)
				switch event.Kind {
				case core.KindTextNote:
					log.Println("[\033[32m*\033[0m] Relay")
					log.Printf("  CreatedAt: %d", event.CreatedAt)
					log.Printf("  Content: %s", event.Content)
				case core.KindSetMetadata:
					log.Println("[\033[32m*\033[0m] Relay")
					p, err := core.ParseMetadata(event.Event)
					if err != nil {
						log.Fatalf("unable to pull profile: %#v", err)
					}
					log.Printf("  name: %s", p.Name)
					log.Printf("  about: %s", p.About)
					log.Printf("  picture: %s", p.Picture)
				}
			case "REQ":
				log.Printf("\n[Relay Response] REQ - %v", msg)
			case "OK":
				e := msg.(*core.MessageResult)
				log.Printf("[\033[32m*\033[0m] Relay")
				log.Printf("  status: OK")
				log.Printf("  message: %s", e.Message)
			default:
				log.Fatalln("unknown message type from RELAY")
			}
		}
	}()
	wg.Wait()

	return nil
}
