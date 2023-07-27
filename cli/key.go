package cli

import (
	"flag"
	"log"

	"github.com/rubenvanstaden/nostr"
)

func NewKey(cc *Connection) *Key {

	gc := &Key{
		fs: flag.NewFlagSet("key", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.BoolVar(&gc.gen, "gen", false, "event text note of Kind 1")
	gc.fs.StringVar(&gc.decode, "decode", "", "event JSON to be signed")

	return gc
}

type Key struct {
	fs *flag.FlagSet
	cc *Connection

	// Content of text note
	gen bool
	decode string
}

func (g *Key) Name() string {
	return g.fs.Name()
}

func (g *Key) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Key) Run() error {

	if s.decode != "" {
        pubkey, err := nostr.DecodeBech32(s.decode)
		if err != nil {
			log.Fatal("unable to generate public key")
		}
		log.Printf("%s", pubkey)
    }

	if s.gen {
		sk := nostr.GeneratePrivateKey()
		pk, err := nostr.GetPublicKey(sk)
		if err != nil {
			log.Fatal("unable to generate public key")
		}
        ns, err := nostr.EncodePrivateKey(sk)
		if err != nil {
			log.Fatal("unable to generate public key")
		}
        np, err := nostr.EncodePublicKey(pk)
		if err != nil {
			log.Fatal("unable to generate public key")
		}
		log.Printf("nsec: %s", ns)
		log.Printf("npub: %s", np)
	}

	return nil
}
