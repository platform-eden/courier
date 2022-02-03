package client

import "fmt"

type MessagingClientError struct {
	Method string
	Err    error
}

func (err *MessagingClientError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type NodeIdGenerationError struct {
	Method string
	Err    error
}

func (err *NodeIdGenerationError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type UnregisteredSubscriberError struct {
	Method  string
	Subject string
}

func (err *UnregisteredSubscriberError) Error() string {
	return fmt.Sprintf("%s: no subscribers registered for subject %s", err.Method, err.Subject)
}

type UnknownNodeEventError struct {
	Event string
	Id    string
}

func (err *UnknownNodeEventError) Error() string {
	return fmt.Sprintf("failed handling node event for %s due to unknown event: %s", err.Id, err.Event)
}

type UnregisteredResponseError struct {
	Method    string
	MessageId string
}

func (err *UnregisteredResponseError) Error() string {
	return fmt.Sprintf("%s: no response exists with id %s", err.Method, err.MessageId)
}

type ContextDoneUnsentMessageError struct {
	MessageId string
}

func (err *ContextDoneUnsentMessageError) Error() string {
	return fmt.Sprintf("context completed before message %s could be sent", err.MessageId)
}

type ClientNodeDialError struct {
	Method   string
	Hostname string
	Port     string
	Err      error
}

func (err *ClientNodeDialError) Error() string {
	return fmt.Sprintf("%s: could not create connection at %s:%s: %s", err.Method, err.Hostname, err.Port, err.Err)
}

type ClientNodeSendError struct {
	Method string
	Err    error
}

func (err *ClientNodeSendError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type ClientNodeMessageTypeError struct {
	Method string
	Type   string
}

func (err *ClientNodeMessageTypeError) Error() string {
	return fmt.Sprintf("%s: message must be of type %s", err.Method, err.Type)
}
