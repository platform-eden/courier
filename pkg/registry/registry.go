package registry

import (
	"context"
	"fmt"
	"sync"
)

type NodeMapper interface {
	Node(string) (Node, bool)
	Nodes() map[string]Node
	addNode(Node)
	removeNode(string)
}

type SubscriberLister interface {
	SubscribeToEvents() chan NodeEvent
	closeListeners()
	forwardEvent(event NodeEvent)
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
			registry.closeListeners()
			return
		case e := <-registry.eventIn:
			switch e.Event {
			case Add:
				registry.addNode(e.Node)
			case Remove:
				registry.removeNode(e.Node.Id)
			case Failed:
				// need to add heartbeat service to this
				registry.removeNode(e.Node.Id)
			default:
				errorChannel <- fmt.Errorf("RegisterNodes: %w", &UnknownNodeEventError{
					nodeId:    e.Id,
					eventType: e.Event,
				})
			}

			registry.forwardEvent(e)
		}
	}
}
