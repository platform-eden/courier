package server

import "fmt"

type messagingServerStartError struct {
	Method string
	Err    error
}

func (err *messagingServerStartError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type ChannelSubscriptionError struct {
	Method string
	Err    error
}

func (err *ChannelSubscriptionError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type UnregisteredChannelSubjectError struct {
	Method  string
	Subject string
}

func (err *UnregisteredChannelSubjectError) Error() string {
	return fmt.Sprintf("%s: no channels registered for subject %s", err.Method, err.Subject)
}
