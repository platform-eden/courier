package messagehandler

import "github.com/platform-edn/courier/message"

// A Message that can be managed by our priority queue
type pqMessage struct {
	message.Message
	Priority int // The priority of the item in the queue.
	Index    int // The index of the item in the heap.
}

// takes a regualr message and returns a message that can be used in a priority queue
func NewPQMessage(m message.Message) *pqMessage {
	pqm := pqMessage{
		Message:  m,
		Priority: int(m.Type),
	}

	return &pqm
}
