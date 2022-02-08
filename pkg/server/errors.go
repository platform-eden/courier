package server

import "fmt"

type UnregisteredChannelSubjectError struct {
	Subject string
}

func (err *UnregisteredChannelSubjectError) Error() string {
	return fmt.Sprintf("no channels registered for subject %s", err.Subject)
}

type ForwardMessagingError struct {
	Message string
	Subject string
}

func (err *ForwardMessagingError) Error() string {
	return fmt.Sprintf("did not complete forwarding messages on %s for message %s", err.Subject, err.Message)
}
