package message

import (
	"github.com/google/uuid"
)

// Message to be sent between nodes
type Message struct {
	Id      string
	Type    messageType
	Subject string
	Content []byte
}

// NewPubMessage creates a new publish message
func NewPubMessage(subject string, content []byte) Message {
	m := Message{
		Id:      uuid.New().String(),
		Type:    PubMessage,
		Subject: subject,
		Content: content,
	}

	return m
}

// NewReqMessage creates a new request message
func NewReqMessage(subject string, content []byte) Message {
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
