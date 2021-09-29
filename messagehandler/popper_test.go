package messagehandler

import (
	"testing"
	"time"

	"github.com/platform-edn/courier/message"
)

func TestNewQueuePopper(t *testing.T) {
	pq := newPriorityQueue()

	m1 := pqMessage{
		Message:  message.NewPubMessage("test", []byte("test")),
		Priority: int(message.PubMessage),
	}

	//increase queue size
	pq.safePush(&m1)
	mp, err := newQueuePopper(pq)
	if err == nil {
		t.Fatalf("queuePopper creation should fail when length is greater than 0.  length was: %v", mp.messageQueue.Len())
	}

	//return queue size to 0
	pq.safePop()
	_, err = newQueuePopper(pq)
	if err != nil {
		t.Fatalf("expected newQueuePopper to pass but got error: %s", err)
	}
}

func TestQueuePopper_Subscribe(t *testing.T) {
	pq := newPriorityQueue()

	mp, err := newQueuePopper(pq)
	if err != nil {
		t.Fatalf("could not create message popper: %s", err)
	}

	tc := mp.subscribe("test")

	mp.start()

	go func() {
		m1 := pqMessage{
			Message:  message.NewPubMessage("test", []byte("test")),
			Priority: int(message.PubMessage),
		}

		pq.safePush(&m1)
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

func TestQueuePopper_Send(t *testing.T) {
	pq := newPriorityQueue()

	mp, err := newQueuePopper(pq)
	if err != nil {
		t.Fatalf("could not create message popper: %s", err)
	}

	m1 := message.NewPubMessage("test", []byte("test"))

	err = mp.send(m1)
	if err == nil {
		t.Fatalf("expected sending on nonexistent subject to fail but it passed")
	}

}
