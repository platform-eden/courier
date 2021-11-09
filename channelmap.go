package courier

import (
	"fmt"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/message"
)

type channelMapper interface {
	Add(string) <-chan message.Message
	Subscriptions(string) ([]chan message.Message, error)
	Close()
}

type ChannelMap struct {
	SubjectChannels map[string][]chan message.Message
	Lock            lock.Locker
}

func newChannelMap() *ChannelMap {
	c := ChannelMap{
		SubjectChannels: map[string][]chan message.Message{},
		Lock:            lock.NewTicketLock(),
	}

	return &c
}

func (c *ChannelMap) Add(subject string) <-chan message.Message {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	newchan := make(chan message.Message)
	c.SubjectChannels[subject] = append(c.SubjectChannels[subject], newchan)

	return newchan
}

func (c *ChannelMap) Subscriptions(subject string) ([]chan message.Message, error) {
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
