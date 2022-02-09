package eventer

import (
	"github.com/platform-edn/courier/pkg/proto"
	"github.com/platform-edn/courier/pkg/registry"
)

type EventTypeMap map[proto.NodeEventType]registry.NodeEventType

func NewEventTypeMap() EventTypeMap {
	events := EventTypeMap{
		proto.NodeEventType_Added:   registry.Add,
		proto.NodeEventType_Removed: registry.Remove,
		proto.NodeEventType_Failed:  registry.Failed,
	}

	return events
}
