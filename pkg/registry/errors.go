package registry

import "fmt"

type UnknownNodeEventError struct {
	eventType NodeEventType
	nodeId    string
}

func (err *UnknownNodeEventError) Error() string {
	return fmt.Sprintf("unknown event for %s: %s", err.nodeId, err.eventType)
}
