package message

import (
	"github.com/google/uuid"
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

// creates a new publish message
func NewPubMessage(subject string, content []byte) Message {
	m := Message{
		Id:      uuid.New().String(),
		Type:    PubMessage,
		Subject: subject,
		Content: content,
	}

	return m
}

// creates a new request message
func NewReqMessage(subject string, content []byte) Message {
	m := Message{
		Id:      uuid.New().String(),
		Type:    ReqMessage,
		Subject: subject,
		Content: content,
	}

	return m
}

// creates a new response message
func NewRespMessage(id string, subject string, content []byte) Message {
	m := Message{
		Id:      id,
		Type:    RespMessage,
		Subject: subject,
		Content: content,
	}

	return m
}
