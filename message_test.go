package nostr

import (
	"encoding/json"
	"testing"

	"github.com/rubenvanstaden/test"
)

func TestUnit_MessageEvent(t *testing.T) {
	cases := []struct {
		msg       string
		wantId    int
		wantSubId string
	}{
		{
			msg:       `["EVENT","_",{"id":"dc90c95f09947507c1044e8f48bcf6350aa6bff1507dd4acfc755b9239b5c962","created_at":1644271588,"kind":1,"content":"ping"}]`,
			wantId:    64,
			wantSubId: "_",
		},
		{
			msg:       `["EVENT",{"id":"dc90c95f09947507c1044e8f48bcf6350aa6bff1507dd4acfc755b9239b5c962","created_at":1644271588,"kind":1,"content":"ping"}]`,
			wantId:    64,
			wantSubId: "",
		},
	}

	for _, c := range cases {

		var msg MessageEvent

		err := json.Unmarshal([]byte(c.msg), &msg)
		test.Ok(t, err)
		test.Equals(t, c.wantId, len(msg.Id))
		test.Equals(t, c.wantSubId, msg.GetSubId())

		msgJson, err := json.Marshal(msg)
		test.Ok(t, err)
		test.Equals(t, c.msg, string(msgJson))
	}
}

func TestUnit_MessageReq(t *testing.T) {
	cases := []struct {
		msg        string
		wantFilter int
		wantIds    int
		wantKinds  int
	}{
		{
			msg:        `["REQ","0",[{"ids":["a","b"],"kinds":[1]}]]`,
			wantFilter: 1,
			wantIds:    2,
			wantKinds:  1,
		},
	}

	for _, c := range cases {

		var msg MessageReq

		err := json.Unmarshal([]byte(c.msg), &msg)
		test.Ok(t, err)
		test.Equals(t, c.wantFilter, len(msg.Filters))
		test.Equals(t, c.wantIds, len(msg.Filters[0].Ids))
		test.Equals(t, c.wantKinds, len(msg.Filters[0].Kinds))

		t.Log(msg)

		msgJson, err := json.Marshal(msg)
		test.Ok(t, err)
		test.Equals(t, c.msg, string(msgJson))

		t.Log(string(msgJson))
	}
}
