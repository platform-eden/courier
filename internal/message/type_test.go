package message

import "testing"

func TestMessageType_String(t *testing.T) {
	ts := []string{
		"PubMessage",
		"RespMessage",
		"ReqMessage",
	}

	types := []MessageType{PubMessage, ReqMessage, RespMessage}

	for i, m := range types {
		if m.String() != ts[i] {
			t.Fatalf("Expected string to be %s but got %s", ts[i], m)
		}
	}
}
