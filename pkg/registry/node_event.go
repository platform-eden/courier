package registry

type NodeEventType int

const (
	Add NodeEventType = iota
	Remove
	Failed
	// used for testing only
	Unknown
)

type NodeEvent struct {
	Node
	Event NodeEventType
}

func NewNodeEvent(node Node, event NodeEventType) NodeEvent {
	nodeEvent := NodeEvent{
		Node:  node,
		Event: event,
	}

	return nodeEvent
}

func (m NodeEventType) String() string {
	types := []string{
		"Add",
		"Remove",
		"Failed",
	}

	return types[int(m)]
}
