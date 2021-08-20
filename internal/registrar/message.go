package registrar

import (
	"github.com/google/uuid"
)

// the types of message that can be created
// assigned based on what kind interaction the sender of the message wants with receiver
type MessageType int

const (
	PubMessage MessageType = iota
	ReqMessage
	RespMessage
)

// a message that is sent between nodes
// each message has a unique id to connect request messages with response messages
// subject describes the content of the message
type Message struct {
	Id      string
	Type    MessageType
	Subject string
	Content []byte
}

// creates a new message
func NewMessage(mt MessageType, subject string, content []byte) Message {
	m := Message{
		Id:      uuid.New().String(),
		Type:    mt,
		Subject: subject,
		Content: content,
	}

	return m
}

// A Message that can be managed by our priority queue
type PQMessage struct {
	Message
	Priority int // The priority of the item in the queue.
	Index    int // The index of the item in the heap.
}
