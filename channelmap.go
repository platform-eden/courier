package courier

import (
	"fmt"
)

type channelMapper interface {
	Add(string) <-chan Message
	Subscriptions(string) ([]chan Message, error)
	Close()
}

type ChannelMap struct {
	SubjectChannels map[string][]chan Message
	Lock            Locker
}

func newChannelMap() *ChannelMap {
	c := ChannelMap{
		SubjectChannels: map[string][]chan Message{},
		Lock:            NewTicketLock(),
	}

	return &c
}

func (c *ChannelMap) Add(subject string) <-chan Message {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	newchan := make(chan Message)
	c.SubjectChannels[subject] = append(c.SubjectChannels[subject], newchan)

	return newchan
}

func (c *ChannelMap) Subscriptions(subject string) ([]chan Message, error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	channels, exist := c.SubjectChannels[subject]
	if !exist {
		return nil, fmt.Errorf("no channels for subjects %s", subject)
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
