package messagehandler

import (
	"testing"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/message"
)

func TestNewPriorityQueue(t *testing.T) {
	pq := newPriorityQueue()

	if pq.Len() != 0 || len(pq.Queue) != 0 {
		t.Fatal("priority queue must be on length 0 when instantiated")
	}
}

func TestPriorityQueue_SafePush(t *testing.T) {
	m := &pqMessage{
		Message: message.Message{
			Id:      uuid.NewString(),
			Type:    message.PubMessage,
			Subject: "test",
			Content: []byte("test"),
		},
	}

	pq := newPriorityQueue()

	pq.safePush(m)

	if pq.Len() != 1 {
		t.Fatalf("expected queue length to be %v but got %v", 1, pq.Len())
	}
}

func TestPriorityQueue_SafePop(t *testing.T) {
	pq := newPriorityQueue()

	_, popped := pq.safePop()

	if popped {
		t.Fatalf("expected popped to be false but got %v", popped)
	}

	m1 := &pqMessage{
		Message: message.Message{
			Id:      uuid.NewString(),
			Type:    message.PubMessage,
			Subject: "test",
			Content: []byte("test"),
		},
		Priority: int(message.PubMessage),
	}

	m2 := &pqMessage{
		Message: message.Message{
			Id:      uuid.NewString(),
			Type:    message.ReqMessage,
			Subject: "test",
			Content: []byte("test"),
		},
		Priority: int(message.ReqMessage),
	}

	m3 := &pqMessage{
		Message: message.Message{
			Id:      uuid.NewString(),
			Type:    message.RespMessage,
			Subject: "test",
			Content: []byte("test"),
		},
		Priority: int(message.RespMessage),
	}

	messageList := []*pqMessage{m2, m3, m1, m2}

	for _, m := range messageList {
		pq.safePush(m)
	}

	m, popped := pq.safePop()
	if !popped {
		t.Fatal("expected popped to be true but got false")
	}

	if m.Message.Type != message.RespMessage {
		t.Fatalf("expected type %v but got %v", message.RespMessage, m.Message.Type)
	}
}

func TestPriorityQueue_Empty(t *testing.T) {
	m := &pqMessage{
		Message: message.Message{
			Id:      uuid.NewString(),
			Type:    message.PubMessage,
			Subject: "test",
			Content: []byte("test"),
		},
	}

	pq := newPriorityQueue()

	pq.safePush(m)

	pq.empty()

	if len(pq.Queue) != 0 {
		t.Fatalf("expected queue length to be 0 bit got %v", len(pq.Queue))
	}
}
