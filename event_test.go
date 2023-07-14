package nostr

import (
	"encoding/json"
	"testing"

	"github.com/rubenvanstaden/test"
)

func TestEventParsing(t *testing.T) {
	rawEvents := []string{
		`{"id":"dc90c95f09947507c1044e8f48bcf6350aa6bff1507dd4acfc755b9239b5c961","created_at":1644271588,"kind":1,"content":"ping"}`,
		`{"id":"dc90c95f09947507c1044e8f48bcf6350aa6bff1507dd4acfc755b9239b5c962","created_at":1644271588,"kind":1,"content":"pong"}`,
	}

	for _, raw := range rawEvents {

		var event Event

		err := json.Unmarshal([]byte(raw), &event)
		if err != nil {
			t.Errorf("Failed to parse event json: %v", err)
		}

		got, err := json.Marshal(event)
		if err != nil {
			t.Errorf("Failed to re marshal event as json: %v", err)
		}

		test.Equals(t, raw, string(got))
	}
}

func TestEventSerialization(t *testing.T) {

	cases := []Event{
		{
			Id:        "92570b321da503eac8014b23447301eb3d0bbdfbace0d11a4e4072e72bb7205d",
			CreatedAt: Timestamp(1671028682),
			Kind:      1,
			Content:   "ping",
		},
	}

	for _, event := range cases {

		b, err := json.Marshal(event)
		if err != nil {
			t.Log(event)
			t.Error("failed to serialize this event")
		}

		var got Event
		if err := json.Unmarshal(b, &got); err != nil {
			t.Log(string(b))
			t.Error("failed to re parse event just serialized")
		}

		test.Equals(t, event, got)
	}
}
