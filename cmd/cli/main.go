package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rubenvanstaden/env"
	"github.com/rubenvanstaden/nostr/core"
)

var (
	PRIVATE_KEY = env.String("NSEC")
	PUBLIC_KEY  = env.String("NPUB")
	CONFIG_PATH = env.String("CONFIG_PATH")
)

type Runner interface {
	Init([]string) error
	Run() error
	Name() string
}

func root(args []string, cfg *core.Config, cc *Connection) error {
	if len(args) < 1 {
		return errors.New("you must pass a sub-command")
	}

	cmds := []Runner{
		NewProfile(cfg, cc),
		NewEvent(cc),
		NewFollow(cc),
	}

	subcommand := os.Args[1]

	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
			cmd.Init(os.Args[2:])
			return cmd.Run()
		}
	}

	return fmt.Errorf("unknown subcommand: %s", subcommand)
}

func main() {

	flag.Parse()
	log.SetFlags(0)

	cfg, err := core.DecodeConfig(CONFIG_PATH)
	if err != nil {
		log.Fatalf("unable to decode local cfg: %v", err)
	}

	cc := NewConnection(cfg.Relays[0])
	defer cc.Close()

	// Parse CLI commands and process events
	err = root(os.Args[1:], cfg, cc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
