package core

import (
	"encoding/json"
	"testing"

	"github.com/rubenvanstaden/test"
)

func TestUnit_MessageEvent(t *testing.T) {
	cases := []struct{
        msg string
        wantId int
        wantSubId string
    }{
        {
            msg: `["EVENT","_",{"id":"dc90c95f09947507c1044e8f48bcf6350aa6bff1507dd4acfc755b9239b5c962","created_at":1644271588,"kind":1,"content":"ping"}]`,
            wantId: 64,
            wantSubId: "_",
        },
        {
            msg: `["EVENT",{"id":"dc90c95f09947507c1044e8f48bcf6350aa6bff1507dd4acfc755b9239b5c962","created_at":1644271588,"kind":1,"content":"ping"}]`,
            wantId: 64,
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

