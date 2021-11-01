package mocks

import (
	"time"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type MockObserver struct {
	UpdateCounter     int
	BlackListCounter  int
	Interval          time.Duration
	FailedConnections chan node.Node
	NewNodes          chan (map[string]node.Node)
	Lock              lock.Locker
}

func NewMockObserver(interval time.Duration) *MockObserver {
	o := MockObserver{
		UpdateCounter:     0,
		BlackListCounter:  0,
		Interval:          interval,
		FailedConnections: make(chan node.Node),
		NewNodes:          make(chan map[string]node.Node),
		Lock:              lock.NewTicketLock(),
	}

	return &o
}

func (m *MockObserver) ObserverInterval() time.Duration {
	return m.Interval
}
func (m *MockObserver) AttemptUpdatingNodes() {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.UpdateCounter++
}
func (m *MockObserver) NodeChannel() chan (map[string]node.Node) {
	return m.NewNodes
}
func (m *MockObserver) FailedConnectionChannel() chan node.Node {
	return m.FailedConnections
}
func (m *MockObserver) BlackListNode(node.Node) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	m.BlackListCounter++
}

func (m *MockObserver) BlackListCount() int {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	return m.BlackListCounter
}
