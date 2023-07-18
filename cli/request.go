package cli

import (
	"flag"
	"log"

	"github.com/rubenvanstaden/nostr"
)

func NewRequest(cc *Connection) *Request {

	gc := &Request{
		fs: flag.NewFlagSet("request", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.StringVar(&gc.profile, "profile", "", "user public key to add to following list")
	gc.fs.StringVar(&gc.notes, "notes", "", "remove user via public key")

	return gc
}

type Request struct {
	fs *flag.FlagSet
	cc *Connection

	// Request profile metadata
	profile string

	// Request the last 10 text notes from a specific profile.
	notes string
}

func (g *Request) Name() string {
	return g.fs.Name()
}

func (g *Request) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Request) Run() error {

	if s.profile != "" {

		// Decode npub using NIP-19
		pk, err := nostr.DecodeBech32(s.profile)
		if err != nil {
			log.Fatalf("\nunable to decode npub: %#v", err)
		}

		f := nostr.Filter{
			Authors: []string{pk.(string)},
			Kinds:   []uint32{nostr.KindSetMetadata},
		}

		sub, err := s.cc.Subscribe(nostr.Filters{f})
		if err != nil {
			return err
		}

		log.Printf("[\033[1;36m>\033[0m] Profile metadata for %s", s.profile)
		for event := range sub.EventStream {
			profile, err := nostr.ParseMetadata(*event)
			if err != nil {
				log.Fatalf("unable to pull profile: %#v", err)
			}
			PrintJson(profile)
		}
	}

	if s.notes != "" {

		// Decode npub using NIP-19
		pk, err := nostr.DecodeBech32(s.profile)
		if err != nil {
			log.Fatalf("\nunable to decode npub: %#v", err)
		}

		f := nostr.Filter{
			Authors: []string{pk.(string)},
			Kinds:   []uint32{nostr.KindTextNote},
			Limit:   3,
		}

		sub, err := s.cc.Subscribe(nostr.Filters{f})
		if err != nil {
			return err
		}

		log.Printf("[\033[1;36m>\033[0m] Text notes from %s", s.profile)

		for event := range sub.EventStream {
			PrintJson(event)
		}
	}

	return nil
}
