package registrar

import (
	"fmt"
)

type Pusher interface {
	Start()
}

// handles incoming Messages for a priority queue
// instead of potential pushers pushing directly to the queue, the pusher listens on a channel that
// it makes available for other processes
type MessagePusher struct {
	messageQueue    MessagePriorityQueuer
	MesssageChannel chan (Message)
	StopChannel     chan (int)
}

// takes a PriorityQueuer and returns a MessagePusher
// PriorityQueuer must be empty to successfully create the MessagePusher
func NewMessagePusher(mpq MessagePriorityQueuer) (*MessagePusher, error) {
	pqLength := mpq.Len()

	if pqLength != 0 {
		return nil, fmt.Errorf("expected PriorityQuerer's queue length to be zero but got %v", pqLength)
	}

	mp := MessagePusher{
		messageQueue:    mpq,
		MesssageChannel: make(chan Message),
	}

	return &mp, nil
}

// starts a concurrent process that listens for new messages sent through the MessagePusher's channel and
// pushes them into the Priority Queue
func (mp *MessagePusher) Start() {
	go func() {
		for {
			message := <-mp.MesssageChannel
			pqm := PQMessage{
				Message:  message,
				Priority: int(message.Type),
			}

			mp.messageQueue.SafePush(&pqm)
		}
	}()
}
