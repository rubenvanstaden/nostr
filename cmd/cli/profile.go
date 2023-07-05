package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/nostr/core"
)

func NewProfile(cc *Connection) *Profile {

	gc := &Profile{
		fs: flag.NewFlagSet("profile", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.StringVar(&gc.name, "name", "", "event text note of Kind 1")
	gc.fs.StringVar(&gc.about, "about", "", "event text note of Kind 0")
	gc.fs.StringVar(&gc.picture, "picture", "", "event text note of Kind 2")

	return gc
}

type Profile struct {
	fs *flag.FlagSet
	cc *Connection

	name      string
	about  string
	picture string
}

func (g *Profile) Name() string {
	return g.fs.Name()
}

func (g *Profile) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Profile) Run() error {

	if s.name != "" {
		s.update(s.name)
	}

	if s.about != "" {
		log.Fatalln("[about] not implemented")
	}

	if s.picture != "" {
		log.Fatalln("[picture] not implemented")
	}

	return nil
}

// FIXME: Pull latest profile and update the diff.
func (s *Profile) update(name string) error {

	var event core.MessageEvent

    // This is a metadata event type.
	event.Kind = core.KindSetMetadata

    // Tags not implemented
	event.Tags = nil

	// The note is created now.
	event.CreatedAt = core.Now()

    p := core.Profile{
        Name: name,
        About: "",
        Picture: "",
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

	log.Printf("[\033[32m*\033[0m] Client")
	log.Printf("  Profile updated (name: %s)", name)

	// Transmit event message to the spoke that connects to the relays.
	err = s.cc.socket.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}
