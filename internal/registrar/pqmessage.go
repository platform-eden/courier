package registrar

// A Message that can be managed by our priority queue
type PQMessage struct {
	Message
	Priority int // The priority of the item in the queue.
	Index    int // The index of the item in the heap.
}

func NewPQMessage(message Message) *PQMessage {
	pqm := PQMessage{
		Message:  message,
		Priority: int(message.Type),
	}

	return &pqm
}
