package server

import (
	"github.com/platform-edn/courier/pkg/lock"
	"github.com/platform-edn/courier/pkg/messaging"
)

type ChannelMap struct {
	SubjectChannels map[string][]chan messaging.Message
	Lock            lock.Locker
}

func newChannelMap() *ChannelMap {
	c := ChannelMap{
		SubjectChannels: map[string][]chan messaging.Message{},
		Lock:            lock.NewTicketLock(),
	}

	return &c
}

func (c *ChannelMap) AddSubscriber(subject string) <-chan messaging.Message {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	newchan := make(chan messaging.Message)
	c.SubjectChannels[subject] = append(c.SubjectChannels[subject], newchan)

	return newchan
}

func (c *ChannelMap) Subscriptions(subject string) ([]chan messaging.Message, error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	channels, exist := c.SubjectChannels[subject]
	if !exist {
		return nil, &UnregisteredChannelSubjectError{
			Method:  "ChannelMap.Subscriptions",
			Subject: subject,
		}
	}

	return channels, nil
}

func (c *ChannelMap) Close() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	for _, channels := range c.SubjectChannels {
		for _, channel := range channels {
			close(channel)
		}
	}
}
