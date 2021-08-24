package queue

import (
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/platform-eden/courier/internal/message"
)

func TestNewQueuePusher(t *testing.T) {
	pq := newPriorityQueue()

	m1 := PQMessage{
		Message:  message.NewPubMessage([]*message.Participant{}, "test", []byte("test")),
		Priority: int(message.PubMessage),
	}

	// increase size of queue
	pq.safePush(&m1)

	mp, err := newQueuePusher(pq)
	if err == nil {
		t.Fatalf("QueuePusher creation should fail when length is greater than 0.  length was: %v", mp.messageQueue.Len())
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
		log.Fatalf("error creating QueuePusher: %v", err)
	}

	mp.start()

	m1 := message.NewPubMessage([]*message.Participant{}, "test", []byte("test"))
	m2 := message.NewReqMessage([]*message.Participant{}, "test", []byte("test"))
	m3 := message.NewRespMessage(uuid.NewString(), []*message.Participant{}, "test", []byte("test"))

	messageList := []message.Message{m1, m2, m3}

	mchan := mp.pushChannel()

	for _, message := range messageList {
		mchan <- message
	}

	queueLength := mp.messageQueue.safeLen()
	if queueLength != len(messageList) {
		log.Fatalf("incorrect length of queue. Expected %v but got %v", len(messageList), queueLength)
	}

}
