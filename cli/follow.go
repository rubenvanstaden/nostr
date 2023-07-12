package cli

import (
	"context"
	"flag"
	"log"

	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr"
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

// A subscription to follow a specific npub has to make
// a REQ to connected relays with the proper filter.
func (s *Follow) subscribe(npub string) error {

    ctx := context.TODO()

	// Decode npub using NIP-19
	_, pk, err := crypto.DecodeBech32(npub)
	if err != nil {
		log.Fatalf("\nunable to decode npub: %#v", err)
	}

	f := nostr.Filter{
		Authors: []string{pk.(string)},
		Kinds:   []uint32{nostr.KindTextNote},
		Limit:   3,
	}

	log.Printf("[\033[33m*\033[0m] client requests to follow %s...", npub[:20])

    s.cc.Request(ctx, nostr.Filters{f})

	// Streaming reponses from the connected relay.

	return nil
}
