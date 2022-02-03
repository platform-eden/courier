package registry

import "sync"

type NodeEventSubscribers struct {
	channels []chan NodeEvent
	sync.RWMutex
}

func NewNodeEventSubscribers() *NodeEventSubscribers {
	subscribers := NodeEventSubscribers{
		channels: []chan NodeEvent{},
		RWMutex:  sync.RWMutex{},
	}

	return &subscribers
}

func (subscribers *NodeEventSubscribers) SubscribeToEvents() chan NodeEvent {
	subscribers.Lock()
	defer subscribers.Unlock()

	events := make(chan NodeEvent)
	subscribers.channels = append(subscribers.channels, events)

	return events
}

func (subscribers *NodeEventSubscribers) closeListeners() {
	for _, channel := range subscribers.channels {
		close(channel)
	}
}

func (subscribers *NodeEventSubscribers) forwardEvent(event NodeEvent) {
	for _, channel := range subscribers.channels {
		channel <- event
	}
}
