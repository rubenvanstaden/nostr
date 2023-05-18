package core

import (
	"encoding/json"
	"testing"
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

		js, err := json.Marshal(event)
		if err != nil {
			t.Errorf("Failed to re marshal event as json: %v", err)
		}

		if string(js) != raw {
			t.Log(string(js))
			t.Error("JSON serialization broken")
		}
	}
}

func TestEventSerialization(t *testing.T) {

    // 1. Given a set of events
    // 2. Marshal each event to a JSON string.
    // 3. Unmarshal each JSON string to an event.
    // 4. Compare the new event with the original event.


}
