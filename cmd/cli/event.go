package main

import (
	"flag"
	"log"

	"github.com/rubenvanstaden/nostr/core"
)

func NewEvent(cc *Connection) *Event {

	gc := &Event{
		fs: flag.NewFlagSet("event", flag.ContinueOnError),
		cc: cc,
	}

	gc.fs.StringVar(&gc.note, "note", "", "event text note of Kind 1")

	return gc
}

type Event struct {
	fs *flag.FlagSet
	cc *Connection

	// Content of text note
	note string
}

func (g *Event) Name() string {
	return g.fs.Name()
}

func (g *Event) Init(args []string) error {
	return g.fs.Parse(args)
}

func (s *Event) Run() error {

	if s.note != "" {

		e := core.Event{
			Kind:      core.KindTextNote,
			Tags:      nil,
			CreatedAt: core.Now(),
			Content:   s.note,
		}

		status, err := s.cc.Publish(e)
		if status == core.StatusOK {
			log.Printf("[\033[1;32m+\033[0m] Text note published: [status: %s]", status)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
