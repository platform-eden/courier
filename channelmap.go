package courier

import (
	"fmt"
)

type ChannelMap struct {
	SubjectChannels map[string][]chan Message
	Lock            Locker
}

type UnregisteredChannelSubjectError struct {
	Method  string
	Subject string
}

func (err *UnregisteredChannelSubjectError) Error() string {
	return fmt.Sprintf("%s: no channels registered for subject %s", err.Method, err.Subject)
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
