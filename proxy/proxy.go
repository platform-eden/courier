package proxy

import (
	"log"
	"time"

	"github.com/platform-edn/courier/message"
)

type MessageProxy struct {
	PushChannel   chan message.Message
	Subscriptions map[string][]chan (message.Message)
}

func NewMessageProxy(push chan message.Message) *MessageProxy {
	mp := MessageProxy{
		PushChannel:   push,
		Subscriptions: map[string][]chan (message.Message){},
	}

	go mp.start()

	return &mp
}

// Subscribe takes a subject and returns a channel that will forward messages for that subject
func (mp *MessageProxy) Subscribe(subject string) chan message.Message {
	subscription := mp.Subscriptions[subject]

	subscriber := make(chan message.Message)

	subscription = append(subscription, subscriber)

	mp.Subscriptions[subject] = subscription

	// need better logging here
	log.Printf("%v New subscriber for subject: %s\n", time.Now().Format(time.RFC3339), subject)

	return subscriber
}

func (mp *MessageProxy) start() {
	for m := range mp.PushChannel {
		err := send(m, mp.Subscriptions)
		if err != nil {
			log.Printf(err.Error())
		}
	}
}

func send(m message.Message, subscriptions map[string][]chan message.Message) error {
	subscription := subscriptions[m.Subject]

	if len(subscription) == 0 {
		return &emptySubscriptionError{}
	}

	for _, subscriber := range subscription {
		go func(subscriber chan message.Message) {
			subscriber <- m
		}(subscriber)
	}

	return nil
}
