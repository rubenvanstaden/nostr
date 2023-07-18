package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/rubenvanstaden/nostr"
)

func NewEvent(cc *Connection) *Event {

	gc := &Event{
		fs: flag.NewFlagSet("event", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.StringVar(&gc.note, "note", "", "event text note of Kind 1")
	gc.fs.StringVar(&gc.sign, "sign", "", "event JSON to be signed")

	return gc
}

type Event struct {
	fs *flag.FlagSet
	cc *Connection

	// Content of text note
	note string
	sign string
}

func (g *Event) Name() string {
	return g.fs.Name()
}

func (g *Event) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Event) Run() error {

	if s.note != "" {

		e := nostr.Event{
			Kind:      nostr.KindTextNote,
			Tags:      nil,
			CreatedAt: nostr.Now(),
			Content:   s.note,
		}

		ok, err := s.cc.Publish(e)
		if ok != nil {
			log.Printf("[\033[1;32m+\033[0m] Text note published: [status: %s]", ok.Ok)
		}
		if err != nil {
			return err
		}
	}

	if s.sign != "" {

		var event nostr.Event
		err := json.Unmarshal([]byte(s.sign), &event)
		if err != nil {
			log.Fatalf("unable to unmarshal event: %#v", err)
		}

		// Decode npub using NIP-19
		sk, err := nostr.DecodeBech32(PRIVATE_KEY)
		if err != nil {
			log.Fatalf("\nunable to decode npub: %#v", err)
		}

		event.Sign(sk.(string))

		fmt.Println(event)
	}

	return nil
}
