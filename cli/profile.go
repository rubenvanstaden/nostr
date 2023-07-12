package cli

import (
	"encoding/json"
	"flag"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr"
)

func NewProfile(cfg *Config, cc *Connection) *Profile {

	gc := &Profile{
		fs:  flag.NewFlagSet("profile", flag.ContinueOnError),
		cfg: cfg,
		cc:  cc,
	}

	gc.fs.StringVar(&gc.name, "name", "", "event text note of Kind 1")
	gc.fs.StringVar(&gc.about, "about", "", "event text note of Kind 0")
	gc.fs.StringVar(&gc.picture, "picture", "", "event text note of Kind 2")

	gc.fs.BoolVar(&gc.show, "show", false, "event text note of Kind 1")
	gc.fs.BoolVar(&gc.commit, "commit", false, "event text note of Kind 1")

	return gc
}

type Profile struct {
	fs  *flag.FlagSet
	cfg *Config
	cc  *Connection

	// Change the name field in profile.
	name string

	// Change the name about in profile.
	about string

	// Change the name picture of profile
	picture string

	// Show the current state of local profile.
	show bool

	// Commit profile to relays in listed in config.
	commit bool
}

func (g *Profile) Name() string {
	return g.fs.Name()
}

func (g *Profile) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Profile) Run() error {

	if s.show {
		s.view()
	}

	if s.name != "" {
		s.cfg.Profile.Name = s.name
		s.cfg.Encode()
	}

	if s.about != "" {
		s.cfg.Profile.About = s.about
		s.cfg.Encode()
	}

	if s.picture != "" {
		s.cfg.Profile.Picture = s.picture
		s.cfg.Encode()
	}

	// Commit event to relays to update profile.
	if s.commit {
		e := nostr.Event{
			Kind:      nostr.KindSetMetadata,
			Tags:      nil,
			CreatedAt: nostr.Now(),
			Content:   s.cfg.Profile.String(),
		}
		status, err := s.cc.Publish(e)
		log.Printf("Profile commit status: %s", status)
		if err != nil {
			return err
		}
	}

	return nil
}

// View the current state of the profile as defined in CONFIG_PATH
func (s *Profile) view() error {

	config, err := DecodeConfig(s.cfg.Path)
	if err != nil {
		log.Fatalf("unable to decode local config: %v", err)
	}

	PrintJson(config)

	return nil
}

// Publich a REQ to relay to return and EVENT of Kind 0.
func (s *Profile) request(npub string) error {

	// Decode npub using NIP-19
	_, pk, err := crypto.DecodeBech32(npub)
	if err != nil {
		log.Fatalf("\nunable to decode npub: %#v", err)
	}

	f := nostr.Filter{
		Authors: []string{pk.(string)},
		Kinds:   []uint32{nostr.KindSetMetadata},
	}

	var req nostr.MessageReq
	req.SubscriptionId = "follow" + ":" + strconv.Itoa(s.cc.counter)
	req.Filters = nostr.Filters{f}

	// Marshal to a slice of bytes ready for transmission.
	msg, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("\nunable to marshal incoming REQ event: %#v", err)
	}

	log.Printf("[\033[32m*\033[0m] client")
	log.Printf("  request to show user profile (npub: %s...)", npub[:10])

	// Transmit event message to the spoke that connects to the relays.
	err = s.cc.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}
