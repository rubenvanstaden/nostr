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

	"github.com/gorilla/websocket"
)

var addr = flag.String("relay", "", "http service address")

func parseFilters(filename string, filter *core.Filter) {
    
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

	err = json.Unmarshal(bytes, &filter)
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

    subId := ""
    var filter core.Filter
	note := ""
	if len(args) > 0 {
		if args[0] == "req" && len(args) > 1 {
            subId = args[1]
            parseFilters(args[2], &filter)
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

	if subId != "" {

		var req core.MessageReq

		req.SubscriptionId = subId
        req.Filters = core.Filters{filter}

        fmt.Printf("subId: %s, filter: %#v", req.SubscriptionId, filter)

		// Marshal to a slice of bytes ready for transmission.
		msg, err := json.Marshal(req)
		if err != nil {
			log.Fatalln("unable to marchal incoming event")
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

		msgEvent.Id = core.NewEventId()
		msgEvent.Kind = 1
		msgEvent.CreatedAt = core.Now()
		msgEvent.Content = note

		// Marshal to a slice of bytes ready for transmission.
		msg, err := json.Marshal(msgEvent)
		if err != nil {
			log.Fatalln("unable to marchal incoming event")
		}

		// Transmit event message to the spoke that connects to the relays.
		err = c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}

	// Start a goroutine for streaming messages from the server
    go func() {
        defer c.Close()
        for {
            _, raw, err := c.ReadMessage()
            if err != nil {
                log.Println("read:", err)
                return
            }

            log.Printf("RAW return: %s", raw)

//             msg := core.DecodeMessage(raw)
//             switch msg.Type() {
//             case "EVENT":
//                 log.Printf("EVENT: %s", msg)
//             default:
//                 log.Fatalln("unknown message type from relay")
//             }
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
