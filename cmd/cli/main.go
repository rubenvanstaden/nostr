package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
)

var addr = flag.String("relay", "", "http service address")

func main() {

    flag.Parse()
	log.SetFlags(0)

	args := flag.Args()

	if *addr == "" {
		log.Fatal("Missing required --relay parameter")
	}

	stream := false
	postMsg := ""
	if len(args) > 0 {
		if args[0] == "stream" {
			stream = true
		} else if args[0] == "post" && len(args) > 1 {
			postMsg = strings.Join(args[1:], " ")
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

	// If "post" command was used, send a message
	if postMsg != "" {
		err = c.WriteMessage(websocket.TextMessage, []byte(postMsg))
		if err != nil {
			log.Println("write:", err)
			return
		}
	}

	// Start a goroutine for streaming messages from the server
	if stream {
		go func() {
			defer c.Close()
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				log.Printf("recv: %s", message)
			}
		}()
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