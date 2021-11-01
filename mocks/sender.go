package mocks

import (
	"context"
	"fmt"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/proto"
)

type MockSender struct {
	PublishCount  int
	RequestCount  int
	ResponseCount int
	Messages      []message.Message
	Fail          bool
	Lock          lock.Locker
}

type CountType int

const (
	Publish CountType = iota
	Request
	Response
)

func NewMockSender() *MockSender {
	s := MockSender{
		PublishCount:  0,
		RequestCount:  0,
		ResponseCount: 0,
		Messages:      []message.Message{},
		Fail:          false,
		Lock:          lock.NewTicketLock(),
	}

	return &s
}

func (s *MockSender) Send(ctx context.Context, m message.Message, client proto.MessageServerClient) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if s.Fail {
		return fmt.Errorf("failed sending message")
	}

	switch m.Type {
	case message.PubMessage:
		s.PublishCount++
		s.Messages = append(s.Messages, m)
	case message.ReqMessage:
		s.RequestCount++
		s.Messages = append(s.Messages, m)
	case message.RespMessage:
		s.ResponseCount++
		s.Messages = append(s.Messages, m)
	}

	return nil
}

func (s *MockSender) Length() int {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	return len(s.Messages)
}

func (s *MockSender) Count(t CountType) int {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	switch t {
	case Publish:
		return s.PublishCount
	case Request:
		return s.RequestCount
	case Response:
		return s.ResponseCount
	default:
		return -1
	}
}
