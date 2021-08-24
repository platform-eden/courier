package queue

import "github.com/platform-eden/courier/internal/message"

// A Message that can be managed by our priority queue
type PQMessage struct {
	message.Message
	Priority int // The priority of the item in the queue.
	Index    int // The index of the item in the heap.
}

// takes a regualr message and returns a message that can be used in a priority queue
func NewPQMessage(message message.Message) *PQMessage {
	pqm := PQMessage{
		Message:  message,
		Priority: int(message.Type),
	}

	return &pqm
}
