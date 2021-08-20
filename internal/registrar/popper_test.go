package registrar

import (
	"testing"
	"time"
)

func TestNewMessagePopper(t *testing.T) {
	mpq := NewMessagePriorityQueue()

	m1 := PQMessage{
		Message:  NewMessage(PubMessage, "pub", []byte("pub")),
		Priority: int(PubMessage),
	}

	//increase queue size
	mpq.SafePush(&m1)
	mp, err := NewMessagePopper(mpq)
	if err == nil {
		t.Fatalf("MessagePopper creation should fail when length is greater than 0.  length was: %v", mp.messageQueue.Len())
	}

	//return queue size to 0
	mpq.SafePop()
	_, err = NewMessagePopper(mpq)
	if err != nil {
		t.Fatalf("expected NewMessagePopper to pass but got error: %s", err)
	}
}

func TestMessagePopper_Subscribe(t *testing.T) {
	mpq := NewMessagePriorityQueue()

	mp, err := NewMessagePopper(mpq)
	if err != nil {
		t.Fatalf("could not create message popper: %s", err)
	}

	tc := mp.Subscribe("test")

	mp.Start()

	go func() {
		m1 := PQMessage{
			Message:  NewMessage(PubMessage, "test", []byte("pub")),
			Priority: int(PubMessage),
		}

		mpq.SafePush(&m1)
	}()

	select {
	case msg := <-tc:
		if msg.Subject != "test" {
			t.Fatalf("expected subject to be test but got %s", msg.Subject)
		}
	case <-time.After(time.Second * 1):
		t.Fatal("timeout waitng on subscribe channel")
	}
}

func TestMessagePopper_Send(t *testing.T) {
	mpq := NewMessagePriorityQueue()

	mp, err := NewMessagePopper(mpq)
	if err != nil {
		t.Fatalf("could not create message popper: %s", err)
	}

	m1 := NewMessage(PubMessage, "test", []byte("pub"))

	err = mp.send(m1)
	if err == nil {
		t.Fatalf("expected sending on nonexistent subject to fail but it passed")
	}

}
