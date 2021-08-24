package queue

import (
	"github.com/google/uuid"
	"github.com/platform-eden/courier/internal/message"
)

type MockPriorityQueue struct {
	Length int
}

func NewMockPriorityQueue() *MockPriorityQueue {
	m := MockPriorityQueue{
		Length: 0,
	}

	return &m
}

func (m *MockPriorityQueue) Len() int {
	return m.Length
}

func (m *MockPriorityQueue) safePush(*PQMessage) {
	m.Length++
}

func (m *MockPriorityQueue) safePop() (*PQMessage, bool) {
	m.Length--
	m1 := PQMessage{
		Message: message.Message{
			Id:      uuid.New().String(),
			Type:    message.PubMessage,
			Subject: "test",
			Content: []byte("test"),
		},
	}

	return &m1, true
}

func (m *MockPriorityQueue) empty() {
	m.Length = 0
}
