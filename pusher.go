package courier

import (
	"fmt"
	"time"
)

type pusher interface {
	start()
	pushChannel() chan Message
}

// handles incoming Messages for a priority queue
// instead of potential pushers pushing directly to the queue, the pusher listens on a channel that
// it makes available for other processes
type queuePusher struct {
	messageQueue    priorityQueuer
	messsageChannel chan (Message)
	StopChannel     chan (int)
}

// takes a priorityQueuer and returns a queuePusher
// priorityQueuer must be empty to successfully create the queuePusher
func newQueuePusher(pq priorityQueuer) (*queuePusher, error) {
	pqLength := pq.Len()

	if pqLength != 0 {
		return nil, fmt.Errorf("expected PriorityQuerer's queue length to be zero but got %v", pqLength)
	}

	mp := queuePusher{
		messageQueue:    pq,
		messsageChannel: make(chan Message),
	}

	return &mp, nil
}

// starts a concurrent process that listens for new messages sent through the queuePusher's channel and
// pushes them into the Priority Queue
func (qp *queuePusher) start() {
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
func (qp *queuePusher) pushChannel() chan Message {
	return qp.messsageChannel
}
