package registrar

import (
	"fmt"

	"go.uber.org/zap"
)

type Popper interface {
	Start()
	Subscribe(string) chan (Message)
	send(Message)
}

// retrieves messages from a PriorityQuerer and sends them to subscribers
// subscribers request channels based on a subject name that they will then receive messages through
// subscriptions are a list of channels subscribers have requested based upon a subject
type MessagePopper struct {
	messageQueue  MessagePriorityQueuer
	Subscriptions map[string][]chan (Message)
}

// creates a new MessagePopper
func NewMessagePopper(pq MessagePriorityQueuer) (*MessagePopper, error) {
	pqLength := pq.Len()

	if pqLength != 0 {
		return nil, fmt.Errorf("expected PriorityQuerer's queue length to be zero but got %v", pqLength)
	}

	mp := MessagePopper{
		messageQueue:  pq,
		Subscriptions: map[string][]chan (Message){},
	}

	return &mp, nil
}

// adds a channel to a list of channels underneath a subject name and returns the channel to the requester
func (mp *MessagePopper) Subscribe(subject string) chan Message {
	subscription := mp.Subscriptions[subject]

	subscriber := make(chan Message)

	subscription = append(subscription, subscriber)

	mp.Subscriptions[subject] = subscription

	return subscriber
}

// signals the MessagePopper to start reading from the PriorityQueuer and send messages to subscribers
func (mp *MessagePopper) Start() {
	go func() {
		for {
			message, popped := mp.messageQueue.SafePop()
			if popped {
				err := mp.send(message.Message)
				if err != nil {
					zap.S().Errorf("sendings message to subscribes failed: %s", err)
				}
			}
		}
	}()
}

type EmptySubscriptionError struct{}

func (m *EmptySubscriptionError) Error() string {
	return "cannot sends message for subject with no subscribers"
}

// gets a messages subject and then sends the message through all of the channels subscribed
func (mp *MessagePopper) send(message Message) error {
	subscription := mp.Subscriptions[message.Subject]

	if len(subscription) == 0 {
		return &EmptySubscriptionError{}
	}

	for _, subscriber := range subscription {
		go func(subscriber chan Message) {
			subscriber <- message
		}(subscriber)
	}

	return nil
}
