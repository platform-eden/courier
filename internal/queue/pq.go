package queue

import (
	"container/heap"
)

type PriorityQueuer interface {
	Len() int
	safeLen() int
	safePop() (*PQMessage, bool)
	safePush(*PQMessage)
	empty()
}

// a priority queue that holds messages
type PriorityQueue struct {
	Queue []*PQMessage
	lock  Locker
}

//creates a new PriorityQueue
func newPriorityQueue() *PriorityQueue {
	pq := PriorityQueue{
		Queue: make([]*PQMessage, 0),
		lock:  newTicketLock(),
	}

	heap.Init(&pq)
	return &pq
}

// gets length of the queue
func (pq *PriorityQueue) Len() int { return len(pq.Queue) }

// needed to implemented heap interface and is used for comparing nodes in queue
func (pq *PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq.Queue[i].Priority > pq.Queue[j].Priority
}

// needed to implemented heap interface and is used for swapping nodes in queue
func (pq *PriorityQueue) Swap(i, j int) {
	pq.Queue[i], pq.Queue[j] = pq.Queue[j], pq.Queue[i]
	pq.Queue[i].Index = i
	pq.Queue[j].Index = j
}

// adds a node into the queue
func (pq *PriorityQueue) Push(message interface{}) {
	n := pq.Len()
	pqMessage := message.(*PQMessage)
	pqMessage.Index = n
	pq.Queue = append(pq.Queue, pqMessage)
}

// pops node from queue
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.Queue
	n := pq.Len()
	pqMessage := old[n-1]
	old[n-1] = nil       // avoid memory leak
	pqMessage.Index = -1 // for safety
	pq.Queue = old[0 : n-1]
	return pqMessage
}

type PQIndexOutOfRangeError struct{}

func (p *PQIndexOutOfRangeError) Error() string {
	return "cannot pop queue with a length of 0"
}

// atomic push
func (pq *PriorityQueue) safePush(message *PQMessage) {
	pq.lock.lock()
	defer pq.lock.unlock()

	heap.Push(pq, message)
}

// atomic check of queue length and pop
func (pq *PriorityQueue) safePop() (*PQMessage, bool) {
	pq.lock.lock()
	defer pq.lock.unlock()

	if pq.Len() == 0 {
		return nil, false
	}

	message := heap.Pop(pq).(*PQMessage)

	return message, true
}

func (pq *PriorityQueue) safeLen() int {
	pq.lock.lock()
	defer pq.lock.unlock()

	return len(pq.Queue)
}

// empties the queue
func (pq *PriorityQueue) empty() {
	pq.lock.lock()
	defer pq.lock.unlock()
	pq.Queue = []*PQMessage{}
}
