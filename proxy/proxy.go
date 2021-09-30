package proxy

import (
	"log"
	"time"

	"github.com/platform-edn/courier/message"
)

type MessageProxy struct {
	PushChannel   chan message.Message
	QuitChannel   chan bool
	Subscriptions map[string][]chan (message.Message)
}

func NewMessageProxy(push chan message.Message, quit chan bool) *MessageProxy {
	mp := MessageProxy{
		PushChannel:   push,
		QuitChannel:   quit,
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
proxy:
	for {
		select {
		case m := <-mp.PushChannel:
			err := send(m, mp.Subscriptions)
			if err != nil {
				log.Printf("%v New subscriber for subject: %s\n", time.Now().Format(time.RFC3339), m.Subject)
			}
		case <-mp.QuitChannel:
			break proxy
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
