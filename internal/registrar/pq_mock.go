package registrar

import "github.com/google/uuid"

type MockMessagePriorityQueue struct {
	Length int
}

func NewMockMessagePriorityQueue() *MockMessagePriorityQueue {
	m := MockMessagePriorityQueue{
		Length: 0,
	}

	return &m
}

func (m *MockMessagePriorityQueue) Len() int {
	return m.Length
}

func (m *MockMessagePriorityQueue) SafePush(*PQMessage) {
	m.Length++
}

func (m *MockMessagePriorityQueue) SafePop() (*PQMessage, bool) {
	m.Length--
	m1 := PQMessage{
		Message: Message{
			Id:      uuid.New().String(),
			Type:    PubMessage,
			Subject: "test",
			Content: []byte("test"),
		},
	}

	return &m1, true
}

func (m *MockMessagePriorityQueue) Empty() {
	m.Length = 0
}
