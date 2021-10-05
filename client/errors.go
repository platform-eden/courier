package client

import "fmt"

type UnregisteredSubjectError struct {
	Subject string
}

// returns an error for the subscribermap if a subject isn't registered
func (e *UnregisteredSubjectError) Error() string {
	return fmt.Sprintf("subject %s is not registered", e.Subject)
}
