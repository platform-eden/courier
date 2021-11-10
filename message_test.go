package courier

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPubMessage(t *testing.T) {
	m := NewPubMessage(uuid.NewString(), "test", []byte("test"))

	if m.Id == "" {
		t.Fatal("expected PubMessage to set id")
	}

	if m.Type != PubMessage {
		t.Fatalf("expected PubMessage but got %s", m.Type)
	}
}

func TestNewReqMessage(t *testing.T) {
	m := NewReqMessage(uuid.NewString(), "test", []byte("test"))

	if m.Id == "" {
		t.Fatal("expected PubMessage to set id")
	}

	if m.Type != ReqMessage {
		t.Fatalf("expected ReqMessage but got %s", m.Type)
	}
}

func TestNewRespMessage(t *testing.T) {
	m := NewRespMessage(uuid.NewString(), "test", []byte("test"))

	if m.Type != RespMessage {
		t.Fatalf("expected ReqMessage but got %s", m.Type)
	}
}

func TestMessageType_String(t *testing.T) {
	ts := []string{
		"PubMessage",
		"RespMessage",
		"ReqMessage",
	}

	types := []messageType{PubMessage, ReqMessage, RespMessage}

	for i, m := range types {
		if m.String() != ts[i] {
			t.Fatalf("Expected string to be %s but got %s", ts[i], m)
		}
	}
}
