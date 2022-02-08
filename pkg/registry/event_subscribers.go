package registry

import "sync"

type NodeEventSubscribers struct {
	Channels []chan NodeEvent
	sync.RWMutex
}

func NewNodeEventSubscribers() *NodeEventSubscribers {
	subscribers := NodeEventSubscribers{
		Channels: []chan NodeEvent{},
		RWMutex:  sync.RWMutex{},
	}

	return &subscribers
}

func (subscribers *NodeEventSubscribers) SubscribeToEvents() chan NodeEvent {
	subscribers.Lock()
	defer subscribers.Unlock()

	events := make(chan NodeEvent)
	subscribers.Channels = append(subscribers.Channels, events)

	return events
}

func (subscribers *NodeEventSubscribers) CloseListeners() {
	for _, channel := range subscribers.Channels {
		close(channel)
	}
}

func (subscribers *NodeEventSubscribers) ForwardEvent(event NodeEvent) {
	for _, channel := range subscribers.Channels {
		channel <- event
	}
}
