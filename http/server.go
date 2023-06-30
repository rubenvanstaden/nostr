package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/rubenvanstaden/nostr/core"
)

const ShutdownTimeout = 1 * time.Second

type Server struct {

	// Vanille http server in golang.
	server *http.Server

	// This router can be easily swapped with something like gofiber.
	router *http.ServeMux

	// Non-TLS address
	addr string

	// Pointer into the spoke hub
	relay *Relay

	// Database store for published events.
	repository core.Repository
}

func NewServer(url string, relay *Relay, repository core.Repository) *Server {

	s := &Server{
		addr:       url,
		relay:      relay,
		server:     &http.Server{},
		router:     http.NewServeMux(),
		repository: repository,
	}

	// Our router is wrapped by another function handler to perform some
	// middleware tasks that cannot be performed by actual middleware.
	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.router.HandleFunc("/", s.handler)

	return s
}

func (s *Server) Addr() string {
	return s.addr
}

// Open validates the server options and begins listening on the bind address.
func (s *Server) Open() error {

	ls, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	// Begin serving requests on the listener. We use Serve() instead of
	// ListenAndServe() because it allows us to check for listen errors (such
	// as trying to use an already open port) synchronously.
	go s.server.ServeTLS(ls, "out/relay.snort.social.crt", "out/relay.snort.social.key")

	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, r)
}
