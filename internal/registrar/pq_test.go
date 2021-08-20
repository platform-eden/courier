package registrar

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewMessagePriorityQueue(t *testing.T) {
	mpq := NewMessagePriorityQueue()

	if mpq.Len() != 0 || len(mpq.Queue) != 0 {
		t.Fatal("priority queue must be on length 0 when instantiated")
	}
}

func TestMessagePriorityQueue_SafePush(t *testing.T) {
	m := &PQMessage{
		Message: Message{
			Id:      uuid.NewString(),
			Type:    PubMessage,
			Subject: "test",
			Content: []byte("test"),
		},
	}

	mpq := NewMessagePriorityQueue()

	mpq.SafePush(m)

	if mpq.Len() != 1 {
		t.Fatalf("expected queue length to be %v but got %v", 1, mpq.Len())
	}
}

func TestMessagePriorityQueue_SafePop(t *testing.T) {
	mpq := NewMessagePriorityQueue()

	_, popped := mpq.SafePop()

	if popped {
		t.Fatalf("expected popped to be false but got %v", popped)
	}

	m1 := &PQMessage{
		Message: Message{
			Id:      uuid.NewString(),
			Type:    PubMessage,
			Subject: "test",
			Content: []byte("test"),
		},
		Priority: int(PubMessage),
	}

	m2 := &PQMessage{
		Message: Message{
			Id:      uuid.NewString(),
			Type:    ReqMessage,
			Subject: "test",
			Content: []byte("test"),
		},
		Priority: int(ReqMessage),
	}

	m3 := &PQMessage{
		Message: Message{
			Id:      uuid.NewString(),
			Type:    RespMessage,
			Subject: "test",
			Content: []byte("test"),
		},
		Priority: int(RespMessage),
	}

	messageList := []*PQMessage{m2, m3, m1, m2}

	for _, m := range messageList {
		mpq.SafePush(m)
	}

	message, popped := mpq.SafePop()
	if !popped {
		t.Fatal("expected popped to be true but got false")
	}

	if message.Message.Type != RespMessage {
		t.Fatalf("expected type %v but got %v", RespMessage, message.Message.Type)
	}
}
