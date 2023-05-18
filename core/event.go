package core

import (
	"encoding/json"
	"log"
)

const (
	KindSetMetadata     uint32 = 0
	KindTextNote        uint32 = 1
	KindRecommendServer uint32 = 2
)

type EventId string

type Event struct {
	Id        EventId   `json:"id"`
	CreatedAt Timestamp `json:"created_at"`
	Kind      uint32    `json:"kind"`
	Content   string    `json:"content"`
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
