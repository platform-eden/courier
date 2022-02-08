package registry

import (
	"context"
	"fmt"
	"sync"
)

type NodeMapper interface {
	AddNode(Node)
	RemoveNode(string)
}

type SubscriberLister interface {
	SubscribeToEvents() chan NodeEvent
	CloseListeners()
	ForwardEvent(event NodeEvent)
}
type NodeRegistry struct {
	eventIn chan NodeEvent
	NodeMapper
	SubscriberLister
}

func NewNodeRegistry() *NodeRegistry {
	r := &NodeRegistry{
		NodeMapper:       NewNodeMap(),
		eventIn:          make(chan NodeEvent),
		SubscriberLister: NewNodeEventSubscribers(),
	}

	return r
}

func (registry NodeRegistry) EventInChannel() chan NodeEvent {
	return registry.eventIn
}

func (registry *NodeRegistry) RegisterNodes(ctx context.Context, wg *sync.WaitGroup, errorChannel chan error) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			registry.CloseListeners()
			return
		case e := <-registry.eventIn:
			switch e.Event {
			case Add:
				registry.AddNode(e.Node)
				registry.ForwardEvent(e)
			case Remove:
				registry.RemoveNode(e.Node.Id)
				registry.ForwardEvent(e)
			case Failed:
				// need to add heartbeat service to this
				registry.RemoveNode(e.Node.Id)
				registry.ForwardEvent(e)
			default:
				errorChannel <- fmt.Errorf("RegisterNodes: %w", &UnknownNodeEventError{
					nodeId:    e.Id,
					eventType: e.Event,
				})
			}
		}
	}
}
