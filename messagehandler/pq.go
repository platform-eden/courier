package messagehandler

import (
	"container/heap"

	"github.com/platform-edn/courier/lock"
)

type priorityQueuer interface {
	Len() int
	safeLen() int
	safePop() (*pqMessage, bool)
	safePush(*pqMessage)
	empty()
}

// a priority queue that holds messages
type priorityQueue struct {
	Queue []*pqMessage
	lock  lock.Locker
}

//creates a new priorityQueue
func newPriorityQueue() *priorityQueue {
	pq := priorityQueue{
		Queue: make([]*pqMessage, 0),
		lock:  lock.NewTicketLock(),
	}

	heap.Init(&pq)
	return &pq
}

// gets length of the queue
func (pq *priorityQueue) Len() int { return len(pq.Queue) }

// needed to implemented heap interface and is used for comparing nodes in queue
func (pq *priorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq.Queue[i].Priority > pq.Queue[j].Priority
}

// needed to implemented heap interface and is used for swapping nodes in queue
func (pq *priorityQueue) Swap(i, j int) {
	pq.Queue[i], pq.Queue[j] = pq.Queue[j], pq.Queue[i]
	pq.Queue[i].Index = i
	pq.Queue[j].Index = j
}

// adds a node into the queue
func (pq *priorityQueue) Push(m interface{}) {
	n := pq.Len()
	pqMessage := m.(*pqMessage)
	pqMessage.Index = n
	pq.Queue = append(pq.Queue, pqMessage)
}

// pops node from queue
func (pq *priorityQueue) Pop() interface{} {
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
func (pq *priorityQueue) safePush(message *pqMessage) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	heap.Push(pq, message)
}

// atomic check of queue length and pop
func (pq *priorityQueue) safePop() (*pqMessage, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.Len() == 0 {
		return nil, false
	}

	message := heap.Pop(pq).(*pqMessage)

	return message, true
}

func (pq *priorityQueue) safeLen() int {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	return len(pq.Queue)
}

// empties the queue
func (pq *priorityQueue) empty() {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	pq.Queue = []*pqMessage{}
}
