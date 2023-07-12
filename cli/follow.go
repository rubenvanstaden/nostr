package cli

import (
	"context"
	"flag"
	"log"

	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr"
)

func NewFollow(cfg *Config, cc *Connection) *Follow {

	gc := &Follow{
		fs: flag.NewFlagSet("follow", flag.ContinueOnError),
        cfg: cfg,
		cc: cc,
	}

	gc.fs.StringVar(&gc.ls, "ls", "", "list all following users")
	gc.fs.StringVar(&gc.add, "add", "", "user public key to add to following list")
	gc.fs.StringVar(&gc.remove, "remove", "", "remove user via public key")

	return gc
}

type Follow struct {
	fs *flag.FlagSet
    cfg *Config
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

    // 1. Update local config with new user.

    s.cfg.Following[npub] = Author{
        PublicKey: npub,
        Name: "Alice",
    }

    // Save to persistent file.
    s.cfg.Encode()

    // 2. Send REQ to relays to add author.

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

    err = s.cc.Request(ctx, nostr.Filters{f})
	if err != nil {
		log.Fatalf("\nunable to request new subsciption npub: %#v", err)
	}

    for event := range s.cc.EventStream {
        PrintJson(event)
    }

	return nil
}
