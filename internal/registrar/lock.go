package registrar

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type Locker interface {
	Lock()
	Unlock()
}

// a ticket lock implementation
// acts as a spinlock and insures goroutines complete in the order they called the lock in
// this helps avoid starvation of goroutines
type TicketLock struct {
	_      sync.Mutex // for copy protection compiler warning
	ticket uint64
	next   uint64
}

// creates a new TicketLock starting at ticket 0
func NewTicketLock() *TicketLock {

	l := TicketLock{
		ticket: 0,
		next:   0,
	}

	return &l
}

// waits until the goroutine's ticket equals the next ticket to be served
func (tl *TicketLock) Lock() {
	t := atomic.AddUint64(&tl.ticket, 1) - 1
	for atomic.LoadUint64(&tl.next) != t {
		runtime.Gosched()
	}
}

// signals a completion of a ticket and makes the lock available for the next ticket
func (tl *TicketLock) Unlock() {
	atomic.AddUint64(&tl.next, 1)
}
