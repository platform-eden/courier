package queue

import (
	"fmt"
	"time"

	"github.com/platform-eden/courier/internal/message"
	"go.uber.org/zap"
)

type Popper interface {
	start()
	subscribe(string) chan message.Message
	send(message.Message) error
}

// retrieves messages from a PriorityQuerer and sends them to subscribers
// subscribers request channels based on a subject name that they will then receive messages through
// subscriptions are a list of channels subscribers have requested based upon a subject
type QueuePopper struct {
	messageQueue  PriorityQueuer
	Subscriptions map[string][]chan (message.Message)
}

// creates a new QueuePopper
func newQueuePopper(pq PriorityQueuer) (*QueuePopper, error) {
	pqLength := pq.Len()

	if pqLength != 0 {
		return nil, fmt.Errorf("expected PriorityQuerer's queue length to be zero but got %v", pqLength)
	}

	mp := QueuePopper{
		messageQueue:  pq,
		Subscriptions: map[string][]chan (message.Message){},
	}

	return &mp, nil
}

// adds a channel to a list of channels underneath a subject name and returns the channel to the requester
func (mp *QueuePopper) subscribe(subject string) chan message.Message {
	subscription := mp.Subscriptions[subject]

	subscriber := make(chan message.Message)

	subscription = append(subscription, subscriber)

	mp.Subscriptions[subject] = subscription

	// need better logging here
	fmt.Printf("%v New subscriber for subject: %s\n", time.Now().Format(time.RFC3339), subject)

	return subscriber
}

// signals the QueuePopper to start reading from the PriorityQueuer and send messages to subscribers
func (mp *QueuePopper) start() {
	fmt.Printf("%v popper started\n", time.Now().Format(time.RFC3339))
	go func() {
		for {
			message, popped := mp.messageQueue.safePop()
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
func (mp *QueuePopper) send(m message.Message) error {
	subscription := mp.Subscriptions[m.Subject]

	if len(subscription) == 0 {
		return &EmptySubscriptionError{}
	}

	for _, subscriber := range subscription {
		go func(subscriber chan message.Message) {
			subscriber <- m
		}(subscriber)
	}

	return nil
}
