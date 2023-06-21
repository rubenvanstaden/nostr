package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// TODO: Split this into DecodeMessageFromRelay and DecodeMessageFromClient.
func DecodeMessage(msg []byte) Message {

	// Extract message label (EVENT, REQ, CLOSE, EOSE, NOTICE) from byte slice.
	firstComma := bytes.Index(msg, []byte{','})
	if firstComma == -1 {
		return nil
	}
	label := msg[0:firstComma]

	var v Message
	switch {
	case bytes.Contains(label, []byte("EVENT")):
		v = &MessageEvent{}
	case bytes.Contains(label, []byte("REQ")):
		v = &MessageReq{}
	default:
		log.Fatalln("cannot decode message")
	}

	if err := v.UnmarshalJSON(msg); err != nil {
		return nil
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

// NIP-01 - ["EVENT", <event JSON as defined above>]
func (s MessageEvent) MarshalJSON() ([]byte, error) {

	msg := append([]byte(nil), []byte(`["EVENT",`)...)

	if len(s.SubscriptionId) != 0 {
		msg = append(msg, []byte(s.SubscriptionId+`,`)...)
	}

	//bytes, err := json.Marshal(s.Event)
	//if err != nil {
	//		log.Fatal(err)
	//	}

	bytes := s.Serialize()

	msg = append(msg, bytes...)
	msg = append(msg, []byte(`]`)...)

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

// NIP-01 - ["REQ", <subscription_id>, <filters JSON>...]
func (s MessageReq) MarshalJSON() ([]byte, error) {

	msg := []byte(nil)

	// Open message array.
	msg = append(msg, []byte(`[`)...)

	// Add message label
	msg = append(msg, []byte(`"REQ",`)...)

	// Add subscription ID between string braces.
	msg = append(msg, []byte(`"`+s.SubscriptionId+`",`)...)

	// Open the filter list
	msg = append(msg, []byte(`[`)...)

	for i, v := range s.Filters {

		// Encode the individual filter.
		bytes, err := json.Marshal(v)
		if err != nil {
			log.Fatal(err)
		}

		// Add filter data to json list.
		msg = append(msg, bytes...)

		// Add delimiter to next item, except the last one
		if i != len(s.Filters)-1 {
			msg = append(msg, []byte(`,`)...)
		}
	}

	// Close the filter list
	msg = append(msg, []byte(`]`)...)

	// Close the entire message.
	msg = append(msg, []byte(`]`)...)

	return msg, nil
}
