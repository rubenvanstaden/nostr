package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"

	"crypto/x509"
	"io"

	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/rubenvanstaden/crypto"
	"github.com/rubenvanstaden/env"
	"github.com/rubenvanstaden/nostr/core"
)

var addr = flag.String("relay", "", "relay websocket address")

var (
	PRIVATE_KEY = env.String("NSEC")
	PUBLIC_KEY  = env.String("NPUB")
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

	subId := ""
	note := ""
	var filters core.Filters

	if len(args) > 0 {
		if args[0] == "req" && len(args) > 1 {
			subId = args[1]
			parseFilters(args[2], &filters)
		} else if args[0] == "note" && len(args) > 1 {
			note = strings.Join(args[1:], " ")
		}
	}

	// Load our CA certificate
	pemCerts, err := ioutil.ReadFile("out/relay.damus.io.pem")
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(pemCerts) {
		log.Fatal("Couldn't append certs")
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	tlsConfig.BuildNameToCertificate()

	// Connect to WebSocket server
	u := url.URL{Scheme: "wss", Host: *addr, Path: "/wss"}
	log.Printf("connecting to %s", u.String())

    // Configure our dialer to use our custom HTTP client
	d := websocket.Dialer{
		TLSClientConfig: tlsConfig,
	}

	// Connect to the WebSocket server
	c, _, err := d.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

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

		msgEvent.Tags = nil

		// The note is created now.
		msgEvent.CreatedAt = core.Now()

		// The user note that should be trimmed properly.
		msgEvent.Content = note

		// Apply NIP-19 to decode user-friendly secrets.
		var sk string
		if _, s, e := crypto.DecodeBech32(PRIVATE_KEY); e == nil {
			sk = s.(string)
		}
		if pub, e := crypto.GetPublicKey(sk); e == nil {
			msgEvent.PubKey = pub
			if npub, e := crypto.EncodePublicKey(pub); e == nil {
				fmt.Fprintln(os.Stderr, "using:", npub)
			}
		}

		log.Printf("sk: %s", sk)
		log.Printf("pk: %s", msgEvent.PubKey)

		// Set public with which the event wat pushed.
		//msgEvent.PubKey = pk

		// We have to sign last, since the signature is dependent on the event content.
		msgEvent.Sign(sk)

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

	// Streaming reponses from the connected relay.
	go func() {
		defer c.Close()
		for {
			_, raw, err := c.ReadMessage()
			if err != nil {
				log.Fatalln(err)
				return
			}

			log.Println("\nRelay Response:")
			log.Println(string(raw))

			msg := core.DecodeMessage(raw)
			switch msg.Type() {
			case "EVENT":
				log.Printf("[Relay Response] EVENT: %#v", msg)
			case "REQ":
				log.Printf("[Relay Response] REQ: %#v", msg)
			case "OK":
				log.Printf("[Relay Response] OK: %#v", msg)
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
