package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"noztr/core"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rubenvanstaden/env"
	"github.com/gorilla/websocket"
)

var addr = flag.String("relay", "", "http service address")

var (
    PRIVATE_KEY = env.String("NSEC")
    PUBLIC_KEY = env.String("NPUB")
)

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

	args := flag.Args()

	if *addr == "" {
		log.Fatal("Missing required --relay parameter")
	}

	nsec := ""
	subId := ""
	note := ""
	var filters core.Filters

	if len(args) > 0 {
        // Generate a new private, public key pair.
		if args[0] == "gen" && len(args) == 1 {
			nsec = args[0]
        } else if args[0] == "req" && len(args) > 1 {
			subId = args[1]
			parseFilters(args[2], &filters)
		} else if args[0] == "note" && len(args) > 1 {
			note = strings.Join(args[1:], " ")
		}
	}

	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

    // We want to create an account (sk, pk) pair first, then publish notes.
	if nsec != "" {
        sk := core.GeneratePrivateKey()

        pk, err := core.GetPublicKey(sk)
        if err != nil {
            log.Fatal("unable to generate public key")
        }

        log.Printf("nsec: %s", sk)
        log.Printf("npub: %s", pk)
    }

	if subId != "" {

		var req core.MessageReq
		req.SubscriptionId = subId
		req.Filters = filters

		// Marshal to a slice of bytes ready for transmission.
		msg, err := json.Marshal(req)
		if err != nil {
			log.Fatalf("\nunable to marshal incoming REQ event: %#v", err)
		}

		// Transmit event message to the spoke that connects to the relays.
		err = c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}

	// If "post" command was used, send a message
	if note != "" {

		var msgEvent core.MessageEvent

		msgEvent.Kind = 1

        // The note is created now.
		msgEvent.CreatedAt = core.Now()

        // The user note that should be trimmed properly.
		msgEvent.Content = note

        // Set public with which the event wat pushed.
        msgEvent.PubKey = PUBLIC_KEY

        // We have to sign last, since the signature is dependent on the event content.
        msgEvent.Sign(PRIVATE_KEY)

		// Marshal the signed event to a slice of bytes ready for transmission.
		msg, err := json.Marshal(msgEvent)
		if err != nil {
			log.Fatalln("unable to marchal incoming event")
		}

        log.Println("\nMSG:")
        log.Println(string(msg))

		// Transmit event message to the spoke that connects to the relays.
		err = c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Fatalln(err)
			return
		}
	}

	// Start a goroutine for streaming messages from the server
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
				log.Printf("[Relay Response] EVENT: %s", msg)
			case "REQ":
				log.Printf("[Relay Response] REQ: %s", msg)
			default:
				log.Fatalln("unknown message type from RELAY")
			}
		}
	}()

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
