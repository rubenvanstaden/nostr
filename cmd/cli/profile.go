package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr/core"
)

func NewProfile(cfg *core.Config, cc *Connection) *Profile {

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
	cfg *core.Config
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

	if s.commit {
		s.publish()
	}

	return nil
}

func (s *Profile) publish() error {

	var event core.MessageEvent

	// This is a metadata event type.
	event.Kind = core.KindSetMetadata

	// Tags not implemented
	event.Tags = nil

	// The note is created now.
	event.CreatedAt = core.Now()

	// The user note that should be trimmed properly.
	event.Content = s.cfg.Profile.String()

	// Apply NIP-19 to decode user-friendly secrets.
	var sk string
	if _, s, e := crypto.DecodeBech32(PRIVATE_KEY); e == nil {
		sk = s.(string)
	}
	if pub, e := crypto.GetPublicKey(sk); e == nil {
		// Set public with which the event wat pushed.
		event.PubKey = pub
		if npub, e := crypto.EncodePublicKey(pub); e == nil {
			fmt.Fprintln(os.Stderr, "using:", npub)
		}
	}

	// We have to sign last, since the signature is dependent on the event content.
	event.Sign(sk)

	// Marshal the signed event to a slice of bytes ready for transmission.
	msg, err := json.Marshal(event)
	if err != nil {
		log.Fatalln("unable to marchal incoming event")
	}

	log.Printf("[\033[32m*\033[0m] client")
	log.Printf("  request to update profile")
	log.Printf("    - name: %s", s.cfg.Profile.Name)
	log.Printf("    - about: %s", s.cfg.Profile.About)
	log.Printf("    - picture: %s", s.cfg.Profile.Picture)

	// Transmit event message to the spoke that connects to the relays.
	err = s.cc.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}

// View the current state of the profile as defined in CONFIG_PATH
func (s *Profile) view() error {

	config, err := core.DecodeConfig(CONFIG_PATH)
	if err != nil {
		log.Fatalf("unable to decode local config: %v", err)
	}

	log.Printf("\n%#v\n", config)

	return nil
}

// Publich a REQ to relay to return and EVENT of Kind 0.
func (s *Profile) request(npub string) error {

	// Decode npub using NIP-19
	_, pk, err := crypto.DecodeBech32(npub)
	if err != nil {
		log.Fatalf("\nunable to decode npub: %#v", err)
	}

	f := core.Filter{
		Authors: []string{pk.(string)},
		Kinds:   []uint32{core.KindSetMetadata},
	}

	var req core.MessageReq
	req.SubscriptionId = "follow" + ":" + strconv.Itoa(s.cc.counter)
	req.Filters = core.Filters{f}

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
