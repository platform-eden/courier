package courier

import (
	"fmt"
	"log"
	"time"
)

type popper interface {
	start()
	subscribe(string) chan Message
	send(Message) error
}

// retrieves messages from a PriorityQuerer and sends them to subscribers
// subscribers request channels based on a subject name that they will then receive messages through
// subscriptions are a list of channels subscribers have requested based upon a subject
type queuePopper struct {
	messageQueue  priorityQueuer
	Subscriptions map[string][]chan (Message)
}

// creates a new queuePopper
func newQueuePopper(pq priorityQueuer) (*queuePopper, error) {
	pqLength := pq.Len()

	if pqLength != 0 {
		return nil, fmt.Errorf("expected PriorityQuerer's queue length to be zero but got %v", pqLength)
	}

	mp := queuePopper{
		messageQueue:  pq,
		Subscriptions: map[string][]chan (Message){},
	}

	return &mp, nil
}

// adds a channel to a list of channels underneath a subject name and returns the channel to the requester
func (mp *queuePopper) subscribe(subject string) chan Message {
	subscription := mp.Subscriptions[subject]

	subscriber := make(chan Message)

	subscription = append(subscription, subscriber)

	mp.Subscriptions[subject] = subscription

	// need better logging here
	fmt.Printf("%v New subscriber for subject: %s\n", time.Now().Format(time.RFC3339), subject)

	return subscriber
}

// signals the queuePopper to start reading from the priorityQueuer and send messages to subscribers
func (mp *queuePopper) start() {
	fmt.Printf("%v popper started\n", time.Now().Format(time.RFC3339))
	go func() {
		for {
			m, popped := mp.messageQueue.safePop()
			if popped {
				err := mp.send(m.Message)
				if err != nil {
					log.Printf("sending message to subscribes failed: %s\n", err)
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
func (mp *queuePopper) send(m Message) error {
	subscription := mp.Subscriptions[m.Subject]

	if len(subscription) == 0 {
		return &EmptySubscriptionError{}
	}

	for _, subscriber := range subscription {
		go func(subscriber chan Message) {
			subscriber <- m
		}(subscriber)
	}

	return nil
}
