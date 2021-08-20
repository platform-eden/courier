package registrar

import (
	"log"
	"testing"
)

func TestNewMessagePusher(t *testing.T) {
	mpq := NewMessagePriorityQueue()

	m1 := PQMessage{
		Message:  NewMessage(PubMessage, "pub", []byte("pub")),
		Priority: int(PubMessage),
	}

	// increase size of queue
	mpq.SafePush(&m1)

	mp, err := NewMessagePusher(mpq)
	if err == nil {
		t.Fatalf("MessagePusher creation should fail when length is greater than 0.  length was: %v", mp.messageQueue.Len())
	}

	//return size to zero
	mpq.SafePop()
	_, err = NewMessagePusher(mpq)
	if err != nil {
		t.Fatalf("expected NewMessagePusher to pass but got error: %s", err)
	}
}

func TestMessagePusher_Listen(t *testing.T) {
	mpq := NewMockMessagePriorityQueue()

	mp, err := NewMessagePusher(mpq)
	if err != nil {
		log.Fatalf("error creating MessagePusher: %v", err)
	}

	mp.Start()

	m1 := NewMessage(PubMessage, "pub", []byte("pub"))
	m2 := NewMessage(ReqMessage, "req", []byte("req"))
	m3 := NewMessage(RespMessage, "resp", []byte("resp"))

	messageList := []Message{m1, m2, m3}

	for _, message := range messageList {
		mp.MesssageChannel <- message
	}

	queueLength := mp.messageQueue.Len()
	if queueLength != len(messageList) {
		log.Fatalf("incorrect length of queue. Expected %v but got %v", len(messageList), queueLength)
	}

}
