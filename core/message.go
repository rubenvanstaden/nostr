package core

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func DecodeMessage(data []byte) Message {

    var tmp []json.RawMessage

    err := json.Unmarshal(data, &tmp)
    if err != nil {
        log.Fatalln("unable to unmarshal client msg")
    }

	s := strings.Trim(string(tmp[0]), "\"")

    var v Message

	switch s {
	case "EVENT":
		v = &MessageEvent{}
    default:
        log.Fatalln("Cannot decode message")
	}

    return v
}

type MessageType string

type Message interface {

    // Return the message type.
    Type() MessageType

    // Implement json.Unmarshaler interface
	UnmarshalJSON([]byte) error

    // Implement json.Marshaler interface
	MarshalJSON() ([]byte, error)
}

type Filter struct {
    Ids []EventId
}

// ----------------------------------------------

type MessageEvent struct {
    SubscriptionId string
    Event
}

func (s MessageEvent) GetSubId() string {
    return strings.Trim(s.SubscriptionId, "\"")
}

func (s MessageEvent) Type() MessageType {
    return MessageType("EVENT")
}

func (s *MessageEvent) UnmarshalJSON(data []byte) error {

    var tmp []json.RawMessage

    err := json.Unmarshal(data, &tmp)
    if err != nil {
        log.Fatalln("unable to unmarshal client msg")
    }

    switch len(tmp) {
    case 2:
        return json.Unmarshal(tmp[1], &s.Event)
    case 3:
        s.SubscriptionId = string(tmp[1])
        return json.Unmarshal(tmp[2], &s.Event)
    default:
        return fmt.Errorf("failed to decode EVENT message")
    }
}

func (s MessageEvent) MarshalJSON() ([]byte, error) {

    start := []byte(`["EVENT"`)
    delimter := []byte(`,`)
    end := []byte(`]`)

	jsonData, err := json.Marshal(s.Event)
	if err != nil {
		log.Fatal(err)
	}

    msg := append([]byte(nil), start...)
    msg = append(msg, delimter...)

    if len(s.SubscriptionId) != 0 {
        msg = append(msg, []byte(s.SubscriptionId)...)
        msg = append(msg, delimter...)
    }

    msg = append(msg, jsonData...)
    msg = append(msg, end...)

    return msg, nil
}

