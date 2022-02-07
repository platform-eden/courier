package messaging

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

// Message to be sent between nodes
type Message struct {
	Id      string
	Type    MessageType
	Subject string
	Content []byte
}

// NewPubMessage creates a new publish message
func NewPubMessage(id string, subject string, content []byte) Message {
	m := Message{
		Id:      id,
		Type:    PubMessage,
		Subject: subject,
		Content: content,
	}

	return m
}

// NewReqMessage creates a new request message
func NewReqMessage(id string, subject string, content []byte) Message {
	m := Message{
		Id:      uuid.New().String(),
		Type:    ReqMessage,
		Subject: subject,
		Content: content,
	}

	return m
}

// NewRespMessage creates a new response message
func NewRespMessage(id string, subject string, content []byte) Message {
	m := Message{
		Id:      id,
		Type:    RespMessage,
		Subject: subject,
		Content: content,
	}

	return m
}

func (m MessageType) String() string {
	types := []string{
		"PubMessage",
		"RespMessage",
		"ReqMessage",
	}

	return types[int(m)]
}
