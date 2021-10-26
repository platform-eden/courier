package proxy

import (
	"fmt"
	"log"
	"time"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/message"
)

type ProxyOption func(p *MessageProxy)

func WithMessageChannel(channel chan message.Message) ProxyOption {
	return func(p *MessageProxy) {
		p.InputChannel = channel
	}
}

type MessageProxy struct {
	InputChannel    chan message.Message
	SubscriptionMap map[string][]chan (message.Message)
	Lock            lock.Locker
}

func NewMessageProxy(options ...ProxyOption) (*MessageProxy, error) {
	p := &MessageProxy{
		SubscriptionMap: map[string][]chan (message.Message){},
		Lock:            lock.NewTicketLock(),
	}

	for _, option := range options {
		option(p)
	}

	if p.InputChannel == nil {
		return nil, fmt.Errorf("input channel must be set")
	}

	return p, nil
}

func (p *MessageProxy) MessageChannel() chan message.Message {
	return p.InputChannel
}

func (p *MessageProxy) Subscriptions(subject string) ([]chan (message.Message), error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	subs, exist := p.SubscriptionMap[subject]
	if !exist {
		return nil, fmt.Errorf("no subscribers for subject %s", subject)
	}

	return subs, nil
}

// Subscribe takes a subject and returns a channel that will forward messages for that subject
func (p *MessageProxy) Subscribe(subject string) chan message.Message {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	subscription := p.SubscriptionMap[subject]
	subscriber := make(chan message.Message)
	subscription = append(subscription, subscriber)

	p.SubscriptionMap[subject] = subscription

	log.Printf("%v New subscriber for subject: %s\n", time.Now().Format(time.RFC3339), subject)

	return subscriber
}
