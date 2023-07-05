package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/env"
	"github.com/rubenvanstaden/nostr/core"
)

var (
	RELAY_URL   = env.String("RELAY_URL")
	PRIVATE_KEY = env.String("NSEC")
	PUBLIC_KEY  = env.String("NPUB")
)

type Runner interface {
	Init([]string) error
	Run() error
	Name() string
}

type Connection struct {

	// Web socket connection between client and relay.
	socket *websocket.Conn

	// Counter for subscriptions
	counter int
}

func root(args []string, cc *Connection) error {
	if len(args) < 1 {
		return errors.New("you must pass a sub-command")
	}

	cmds := []Runner{
        NewProfile(cc),
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

func parseFilters(filename string, filters *core.Filters) {

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	err = json.Unmarshal(bytes, &filters)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		os.Exit(1)
	}
}

func main() {

	flag.Parse()
	log.SetFlags(0)

	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: RELAY_URL, Path: ""}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	// Streaming reponses from the connected relay.
	go func() {
		defer c.Close()
		for {
			_, raw, err := c.ReadMessage()
			if err != nil {
				log.Fatalln(err)
				return
			}
			msg := core.DecodeMessage(raw)
			switch msg.Type() {
			case "EVENT":
				event := msg.(*core.MessageEvent)
				log.Println("[\033[32m*\033[0m] Relay")
				log.Printf("  CreatedAt: %d", event.CreatedAt)
				log.Printf("  Content: %s", event.Content)
			case "REQ":
				log.Printf("\n[Relay Response] REQ - %v", msg)
			case "OK":
				e := msg.(*core.MessageResult)
				log.Printf("[\033[32m*\033[0m] Relay")
				log.Printf("  status: OK")
				log.Printf("  message: %s", e.Message)
			default:
				log.Fatalln("unknown message type from RELAY")
			}
		}
	}()

	cc := &Connection{
		socket:  c,
		counter: 0,
	}

	// Parse CLI commands and process events
	err = root(os.Args[1:], cc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Wait for SIGINT (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt

	// Disconnect from the WebSocket server
	log.Println("disconnecting from server")
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}
