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

func NewProfile(cc *Connection) *Profile {

	gc := &Profile{
		fs: flag.NewFlagSet("profile", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.BoolVar(&gc.ls, "ls", false, "event text note of Kind 1")
	gc.fs.StringVar(&gc.name, "name", "", "event text note of Kind 1")
	gc.fs.StringVar(&gc.about, "about", "", "event text note of Kind 0")
	gc.fs.StringVar(&gc.picture, "picture", "", "event text note of Kind 2")

	return gc
}

type Profile struct {
	fs *flag.FlagSet
	cc *Connection

	ls      bool
	name    string
	about   string
	picture string
}

func (g *Profile) Name() string {
	return g.fs.Name()
}

func (g *Profile) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Profile) Run() error {

	if s.ls {
		s.show(PUBLIC_KEY)
	}

	if s.name != "" && s.about != "" && s.picture != "" {
		s.update(s.name, s.about, s.picture)
	}

	return nil
}

// FIXME: Pull latest profile and update the diff.
func (s *Profile) update(name, about, picture string) error {

	var event core.MessageEvent

	// This is a metadata event type.
	event.Kind = core.KindSetMetadata

	// Tags not implemented
	event.Tags = nil

	// The note is created now.
	event.CreatedAt = core.Now()

	p := core.Profile{
		Name:    name,
		About:   about,
		Picture: picture,
	}

	// The user note that should be trimmed properly.
	event.Content = p.String()

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
	log.Printf("    - name: %s", name)
	log.Printf("    - about: %s", about)
	log.Printf("    - picture: %s", picture)

	// Transmit event message to the spoke that connects to the relays.
	err = s.cc.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}

// Publich a REQ to relay to return and EVENT of Kind 0.
func (s *Profile) show(npub string) error {

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
