package queue

import (
	"fmt"
	"time"

	"github.com/platform-eden/courier/internal/message"
)

type Pusher interface {
	start()
	pushChannel() chan message.Message
}

// handles incoming Messages for a priority queue
// instead of potential pushers pushing directly to the queue, the pusher listens on a channel that
// it makes available for other processes
type QueuePusher struct {
	messageQueue    PriorityQueuer
	messsageChannel chan (message.Message)
	StopChannel     chan (int)
}

// takes a PriorityQueuer and returns a QueuePusher
// PriorityQueuer must be empty to successfully create the QueuePusher
func newQueuePusher(pq PriorityQueuer) (*QueuePusher, error) {
	pqLength := pq.Len()

	if pqLength != 0 {
		return nil, fmt.Errorf("expected PriorityQuerer's queue length to be zero but got %v", pqLength)
	}

	mp := QueuePusher{
		messageQueue:    pq,
		messsageChannel: make(chan message.Message),
	}

	return &mp, nil
}

// starts a concurrent process that listens for new messages sent through the QueuePusher's channel and
// pushes them into the Priority Queue
func (qp *QueuePusher) start() {
	fmt.Printf("%v pusher started\n", time.Now().Format(time.RFC3339))
	go func() {
		for {
			message := <-qp.messsageChannel
			pqm := NewPQMessage(message)

			qp.messageQueue.safePush(pqm)
		}
	}()
}

// returns the channel that the pusher receives messages on
func (qp *QueuePusher) pushChannel() chan message.Message {
	return qp.messsageChannel
}
