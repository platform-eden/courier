package messagehandler

import (
	"fmt"

	"github.com/platform-edn/courier/message"
)

type Handler interface {
	PushChannel() chan message.Message
	Subscribe(string) chan message.Message
}

type QueueHandler struct {
	pusher pusher
	popper popper
	queue  priorityQueuer
}

// Returns a queueHandler with a priority queue.  makes a channel avaialble for pushing to the queue and subscribing to subjects
// popped from the queue.
func NewQueueHandler() (*QueueHandler, error) {
	pq := newPriorityQueue()

	pusher, err := newQueuePusher(pq)
	if err != nil {
		return nil, fmt.Errorf("could not create pusher: %s", err)
	}

	popper, err := newQueuePopper(pq)
	if err != nil {
		return nil, fmt.Errorf("could not create popper: %s", err)
	}

	qh := QueueHandler{
		pusher: pusher,
		popper: popper,
		queue:  pq,
	}

	qh.popper.start()
	qh.pusher.start()

	return &qh, nil
}

// returns the channel that the queue handler's pusher receives messages on
func (qh *QueueHandler) PushChannel() chan message.Message {
	return qh.pusher.pushChannel()
}

// takes a subject and returns a channel that will receive messages assigned that subject
func (qh *QueueHandler) Subscribe(subject string) chan message.Message {
	mchan := qh.popper.subscribe(subject)
	return mchan
}
