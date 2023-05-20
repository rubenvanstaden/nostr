package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"strings"
)

type Kind uint32

const (
	KindSetMetadata Kind = 0
	KindTextNote    Kind = 1
)

type EventId string

func NewEventId() EventId {

	// create a slice with a length of 32 bytes
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("unable to generate new event ID: %v", err)
		return ""
	}

	// convert the bytes to a hex string
	return EventId(hex.EncodeToString(b))
}

type Event struct {
	Id        EventId   `json:"id"`
	CreatedAt Timestamp `json:"created_at"`
	Kind      Kind      `json:"kind"`
	Content   string    `json:"content"`
}

func (s Event) GetId() string {
	return strings.Trim(string(s.Id), "\"")
}

func (s Event) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		log.Fatalln("Unable to convert event to string")
	}
	return string(bytes)
}

func (s Event) Serialize() []byte {
	bytes, err := json.Marshal(s)
	if err != nil {
		log.Fatalln("Unable to serialize event")
	}
	return bytes
}
