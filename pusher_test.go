package courier

import (
	"log"
	"testing"

	"github.com/google/uuid"
)

func TestNewQueuePusher(t *testing.T) {
	pq := newPriorityQueue()

	m1 := pqMessage{
		Message:  NewPubMessage("test", []byte("test")),
		Priority: int(PubMessage),
	}

	// increase size of queue
	pq.safePush(&m1)

	mp, err := newQueuePusher(pq)
	if err == nil {
		t.Fatalf("queuePusher creation should fail when length is greater than 0.  length was: %v", mp.messageQueue.Len())
	}

	//return size to zero
	pq.safePop()
	_, err = newQueuePusher(pq)
	if err != nil {
		t.Fatalf("expected newQueuePusher to pass but got error: %s", err)
	}
}

func TestQueuePusher_Listen(t *testing.T) {
	pq := newPriorityQueue()

	mp, err := newQueuePusher(pq)
	if err != nil {
		log.Fatalf("error creating queuePusher: %v", err)
	}

	mp.start()

	m1 := NewPubMessage("test", []byte("test"))
	m2 := NewReqMessage("test", []byte("test"))
	m3 := NewRespMessage(uuid.NewString(), "test", []byte("test"))

	messageList := []Message{m1, m2, m3}

	mchan := mp.pushChannel()

	for _, message := range messageList {
		mchan <- message
	}

	queueLength := mp.messageQueue.safeLen()
	if queueLength != len(messageList) {
		log.Fatalf("incorrect length of queue. Expected %v but got %v", len(messageList), queueLength)
	}

}
