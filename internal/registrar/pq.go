package registrar

import (
	"container/heap"
)

type MessagePriorityQueuer interface {
	Len() int
	SafePop() (*PQMessage, bool)
	SafePush(*PQMessage)
	Empty()
}

// a priority queue that holds messages
type MessagePriorityQueue struct {
	Queue []*PQMessage
	lock  Locker
}

//creates a new MessagePriorityQueue
func NewMessagePriorityQueue() *MessagePriorityQueue {
	mpq := MessagePriorityQueue{
		Queue: make([]*PQMessage, 0),
		lock:  NewTicketLock(),
	}

	heap.Init(&mpq)
	return &mpq
}

// gets length of the queue
func (mpq *MessagePriorityQueue) Len() int { return len(mpq.Queue) }

// needed to implemented heap interface and is used for comparing nodes in queue
func (mpq *MessagePriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return mpq.Queue[i].Priority > mpq.Queue[j].Priority
}

// needed to implemented heap interface and is used for swapping nodes in queue
func (mpq *MessagePriorityQueue) Swap(i, j int) {
	mpq.Queue[i], mpq.Queue[j] = mpq.Queue[j], mpq.Queue[i]
	mpq.Queue[i].Index = i
	mpq.Queue[j].Index = j
}

// adds a node into the queue
func (mpq *MessagePriorityQueue) Push(message interface{}) {
	n := mpq.Len()
	mpqMessage := message.(*PQMessage)
	mpqMessage.Index = n
	mpq.Queue = append(mpq.Queue, mpqMessage)
}

// pops node from queue
func (mpq *MessagePriorityQueue) Pop() interface{} {
	old := mpq.Queue
	n := mpq.Len()
	mpqMessage := old[n-1]
	old[n-1] = nil        // avoid memory leak
	mpqMessage.Index = -1 // for safety
	mpq.Queue = old[0 : n-1]
	return mpqMessage
}

type PQIndexOutOfRangeError struct{}

func (p *PQIndexOutOfRangeError) Error() string {
	return "cannot pop queue with a length of 0"
}

// atomic push
func (mpq *MessagePriorityQueue) SafePush(message *PQMessage) {
	mpq.lock.Lock()
	defer mpq.lock.Unlock()

	heap.Push(mpq, message)
}

// atomic check of queue length and pop
func (mpq *MessagePriorityQueue) SafePop() (*PQMessage, bool) {
	mpq.lock.Lock()
	defer mpq.lock.Unlock()

	if mpq.Len() == 0 {
		return nil, false
	}

	message := heap.Pop(mpq).(*PQMessage)

	return message, true
}

// empties the queue
func (mpq *MessagePriorityQueue) Empty() {
	mpq.lock.Lock()
	defer mpq.lock.Unlock()
	mpq.Queue = []*PQMessage{}
}
