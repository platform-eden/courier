package client

import "fmt"

type MessagingClientError struct {
	Method string
	Err    error
}

func (err *MessagingClientError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type UnregisteredSubjectError struct {
	Subject string
}

func (err *UnregisteredSubjectError) Error() string {
	return fmt.Sprintf("nodes does not have subject %s registered", err.Subject)
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
	MessageId string
}

func (err *UnregisteredResponseError) Error() string {
	return fmt.Sprintf("no response exists with id %s", err.MessageId)
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

type BadMessageTypeError struct {
	Expected string
	Actual   string
}

func (err *BadMessageTypeError) Error() string {
	return fmt.Sprintf("bad message type: expected %s but got %s", err.Expected, err.Actual)
}

type UnregisteredClientNodeError struct {
	Id string
}

func (err *UnregisteredClientNodeError) Error() string {
	return fmt.Sprintf("no client node registered with id %s", err.Id)
}
