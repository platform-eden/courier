package messaging

import "fmt"

type UnregisteredChannelSubjectError struct {
	Method  string
	Subject string
}

func (err *UnregisteredChannelSubjectError) Error() string {
	return fmt.Sprintf("%s: no channels registered for subject %s", err.Method, err.Subject)
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
	Type   messageType
}

func (err *ClientNodeMessageTypeError) Error() string {
	return fmt.Sprintf("%s: message must be of type %s", err.Method, err.Type)
}

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

type NoObserverChannelError struct {
	Method string
}

func (err *NoObserverChannelError) Error() string {
	return fmt.Sprintf("%s: observer channel must be set", err.Method)
}

type SendMessagingServiceMessageError struct {
	Method string
	Err    error
}

func (err *SendMessagingServiceMessageError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type MessagingServiceStartError struct {
	Method string
	Err    error
}

func (err *MessagingServiceStartError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type ExpectedFailureError struct{}

func (err *ExpectedFailureError) Error() string {
	return "failed, but this was expected"
}

type UnregisteredResponseError struct {
	Method    string
	MessageId string
}

func (err *UnregisteredResponseError) Error() string {
	return fmt.Sprintf("%s: no response exists with id %s", err.Method, err.MessageId)
}

type ChannelSubscriptionError struct {
	Method string
	Err    error
}

func (err *ChannelSubscriptionError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type messagingServerStartError struct {
	Method string
	Err    error
}

func (err *messagingServerStartError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type UnregisteredSubscriberError struct {
	Method  string
	Subject string
}

func (err *UnregisteredSubscriberError) Error() string {
	return fmt.Sprintf("%s: no subscribers registered for subject %s", err.Method, err.Subject)
}
