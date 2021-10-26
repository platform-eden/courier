package observer

import (
	"fmt"
	"time"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type StoreObserverOption func(o *StoreObserver)

func WithNodeStorer(store NodeStorer) StoreObserverOption {
	return func(o *StoreObserver) {
		o.store = store
	}
}

func WithObserverInterval(interval time.Duration) StoreObserverOption {
	return func(o *StoreObserver) {
		o.observeInterval = interval
	}
}

func WithSubjects(subjects []string) StoreObserverOption {
	return func(o *StoreObserver) {
		o.subjects = subjects
	}
}

type StoreObserver struct {
	store             NodeStorer
	observeInterval   time.Duration
	currentNodes      *NodeMap
	blackListedNodes  *NodeMap
	nodeChannels      []chan (map[string]node.Node)
	failedConnections chan node.Node
	subjects          []string
	lock              lock.Locker
}

// Returns a paused StoreObserver
func NewStoreObserver(options ...StoreObserverOption) (*StoreObserver, error) {
	s := &StoreObserver{
		observeInterval:   time.Second,
		currentNodes:      NewNodeMap(),
		blackListedNodes:  NewNodeMap(),
		nodeChannels:      []chan (map[string]node.Node){},
		failedConnections: make(chan node.Node),
		subjects:          []string{},
		lock:              lock.NewTicketLock(),
	}

	for _, option := range options {
		option(s)
	}

	if s.store == nil {
		return nil, fmt.Errorf("store observer must be given a store to observe")
	}

	return s, nil
}

// NodeChannel adds a channel to the StoreObserver that will receive a map of Nodes when the NodeStore has updated Nodes and returns it
func (s *StoreObserver) NodeChannel() chan (map[string]node.Node) {
	s.lock.Lock()
	defer s.lock.Unlock()

	channel := make(chan map[string]node.Node)
	s.nodeChannels = append(s.nodeChannels, channel)

	return channel
}

func (s *StoreObserver) ObserverInterval() time.Duration {
	return s.observeInterval
}

// FailedConnectionChannel returns a channel that can receive nodes to blacklist when they have a bad connection
func (s *StoreObserver) FailedConnectionChannel() chan node.Node {
	return s.failedConnections
}

func (s *StoreObserver) GetCourierNodes() ([]*node.Node, error) {
	return s.store.GetSubscribers(s.subjects...)
}

func (s *StoreObserver) TotalBlackListedNodes() int {
	return s.blackListedNodes.Length()
}

func (s *StoreObserver) BlackListedNodes() map[string]node.Node {
	return s.blackListedNodes.Nodes()
}

func (s *StoreObserver) UpdateBlackListedNodes(nodes map[string]node.Node) {
	s.blackListedNodes.Update(nodes)
}

func (s *StoreObserver) AddBlackListNode(n node.Node) {
	s.blackListedNodes.AddNode(n)
}

func (s *StoreObserver) CurrentNodes() map[string]node.Node {
	return s.currentNodes.Nodes()
}

func (s *StoreObserver) RemoveCurrentNode(id string) {
	s.currentNodes.RemoveNode(id)
}

func (s *StoreObserver) UpdateCurrentNodes(nodes map[string]node.Node) {
	s.currentNodes.Update(nodes)
}

func (s *StoreObserver) SendNodes(nodes map[string]node.Node) {
	for _, channel := range s.nodeChannels {
		channel <- nodes
	}
}
