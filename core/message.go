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
	case "REQ":
		v = &MessageReq{}
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
		log.Fatalln("unable to unmarshal EVENT msg")
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

// ----------------------------------------------

type MessageReq struct {
    SubscriptionId string
    Filters
}

func (s MessageReq) GetSubId() string {
	return strings.Trim(s.SubscriptionId, "\"")
}

func (s MessageReq) Type() MessageType {
	return MessageType("REQ")
}

func (s *MessageReq) UnmarshalJSON(data []byte) error {

	var tmp []json.RawMessage

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		log.Fatalln("unable to unmarshal REQ msg")
	}

    s.SubscriptionId = string(tmp[1])
    return json.Unmarshal(tmp[2], &s.Filters)
}

func (s MessageReq) MarshalJSON() ([]byte, error) {

	start := []byte(`["REQ"`)
	delimter := []byte(`,`)
	end := []byte(`]`)

	jsonData, err := json.Marshal(s.Filters)
	if err != nil {
        log.Println("kewferbffbwflekflwkefenlk")
		log.Fatal(err)
	}

    log.Printf("%s", string(jsonData))

	msg := append([]byte(nil), start...)
	msg = append(msg, delimter...)

    msg = append(msg, []byte(s.SubscriptionId)...)
    msg = append(msg, delimter...)

	msg = append(msg, jsonData...)
	msg = append(msg, end...)

	return msg, nil
}

