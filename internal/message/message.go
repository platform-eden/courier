package message

import (
	"github.com/google/uuid"
)

// a message that is sent between nodes
// each message has a unique id to connect request messages with response messages
// subject describes the content of the message
type Message struct {
	Id           string
	Type         MessageType
	Participants []*Participant
	Subject      string
	Content      []byte
}

// creates a new publish message
func NewPubMessage(participants []*Participant, subject string, content []byte) Message {
	m := Message{
		Id:           uuid.New().String(),
		Type:         PubMessage,
		Participants: participants,
		Subject:      subject,
		Content:      content,
	}

	return m
}

// creates a new request message
func NewReqMessage(participants []*Participant, subject string, content []byte) Message {
	m := Message{
		Id:           uuid.New().String(),
		Type:         ReqMessage,
		Participants: participants,
		Subject:      subject,
		Content:      content,
	}

	return m
}

// creates a new response message
func NewRespMessage(id string, participants []*Participant, subject string, content []byte) Message {
	m := Message{
		Id:           id,
		Type:         RespMessage,
		Participants: participants,
		Subject:      subject,
		Content:      content,
	}

	return m
}
