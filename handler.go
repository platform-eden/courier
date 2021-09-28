package courier

import (
	"fmt"
)

type handler interface {
	PushChannel() chan Message
	Subscribe(string) chan Message
}

type queueHandler struct {
	pusher pusher
	popper popper
	queue  priorityQueuer
}

// Returns a queueHandler with a priority queue.  makes a channel avaialble for pushing to the queue and subscribing to subjects
// popped from the queue.
func NewQueueHandler() (*queueHandler, error) {
	pq := newPriorityQueue()

	pusher, err := newQueuePusher(pq)
	if err != nil {
		return nil, fmt.Errorf("could not create pusher: %s", err)
	}

	popper, err := newQueuePopper(pq)
	if err != nil {
		return nil, fmt.Errorf("could not create popper: %s", err)
	}

	qh := queueHandler{
		pusher: pusher,
		popper: popper,
		queue:  pq,
	}

	qh.popper.start()
	qh.pusher.start()

	return &qh, nil
}

// returns the channel that the queue handler's pusher receives messages on
func (qh *queueHandler) PushChannel() chan Message {
	return qh.pusher.pushChannel()
}

// takes a subject and returns a channel that will receive messages assigned that subject
func (qh *queueHandler) Subscribe(subject string) chan Message {
	mchan := qh.popper.subscribe(subject)
	return mchan
}
