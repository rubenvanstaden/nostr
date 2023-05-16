package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketServer(t *testing.T) {

	// Create a test HTTP server with the WebSocket handler
	server := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer server.Close()

	// Convert the server URL to a WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/echo"

	// Create a WebSocket connection
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to establish WebSocket connection: %v", err)
	}
	defer wsConn.Close()

	// Send a message to the WebSocket server
	message := []byte("Hello, WebSocket!")
	err = wsConn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		t.Fatalf("Failed to send message to WebSocket server: %v", err)
	}

	// Read the response from the WebSocket server
	_, response, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response from WebSocket server: %v", err)
	}

	// Verify the response matches the sent message
	if string(response) != string(message) {
		t.Errorf("Unexpected response from WebSocket server. Expected: %s, Got: %s", message, response)
	}
}
