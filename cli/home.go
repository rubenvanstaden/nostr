package cli

import (
	"context"
	"flag"
	"log"

	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr"
)

func NewHome(cfg *Config, cc *Connection) *Home {

	gc := &Home{
		fs: flag.NewFlagSet("home", flag.ContinueOnError),
        cfg: cfg,
		cc: cc,
	}

	gc.fs.BoolVar(&gc.following, "following", false, "event text note of Kind 1")

	return gc
}

type Home struct {
	fs *flag.FlagSet
    cfg *Config
	cc *Connection

	// Content of text note
    following bool
}

func (g *Home) Name() string {
	return g.fs.Name()
}

func (g *Home) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Home) Run() error {

    ctx := context.TODO()

	if s.following {

        // 1. Range over following as defined in local config.

        config, err := DecodeConfig(s.cfg.Path)
        if err != nil {
            log.Fatalf("unable to decode local config: %v", err)
        }

        for _, user := range config.Following {

            _, pk, err := crypto.DecodeBech32(user.PublicKey)
            if err != nil {
                log.Fatalf("\nunable to decode npub: %#v", err)
            }

            // List only the latest 3 event from the author.
            f := nostr.Filter{
                Authors: []string{pk.(string)},
                Kinds:   []uint32{nostr.KindTextNote},
                Limit:   3,
            }

            s.cc.Request(ctx, nostr.Filters{f})

            for event := range s.cc.EventStream {
                PrintJson(event)
            }
        }
	}

	return nil
}
