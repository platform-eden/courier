package testing

import (
	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type MockNodeStore struct {
	lock           lock.Locker
	BroadCastNodes map[string][]*node.Node
	SubscribeNodes map[string][]*node.Node
}

// returns a mock of a Node Store good for testing
func NewMockNodeStore(nodes ...*node.Node) *MockNodeStore {
	bnodes := map[string][]*node.Node{}
	snodes := map[string][]*node.Node{}

	for _, n := range nodes {
		if len(n.SubscribedSubjects) != 0 {
			for _, subject := range n.SubscribedSubjects {
				_, exist := snodes[subject]
				if !exist {
					snodes[subject] = []*node.Node{n}
					continue
				}

				snodes[subject] = append(snodes[subject], n)
			}
		}

		if len(n.BroadcastedSubjects) != 0 {
			for _, subject := range n.BroadcastedSubjects {
				_, exist := bnodes[subject]
				if !exist {
					bnodes[subject] = []*node.Node{n}
					continue
				}

				bnodes[subject] = append(bnodes[subject], n)
			}
		}
	}

	m := MockNodeStore{
		lock:           lock.NewTicketLock(),
		BroadCastNodes: bnodes,
		SubscribeNodes: snodes,
	}

	return &m
}

// Gets all of the subscribers that listen on the given subjects and returns them as a list
func (n *MockNodeStore) GetSubscribers(subjects ...string) ([]*node.Node, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

	nodes := []*node.Node{}
	for _, subject := range subjects {
		subs, exist := n.SubscribeNodes[subject]
		if exist {
			nodes = append(nodes, subs...)
		}
	}
	return nodes, nil
}

// Adds a node to the Node Store.  Should be called when adding a courier service.
func (m *MockNodeStore) AddNode(n *node.Node) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(n.SubscribedSubjects) != 0 {
		for _, subject := range n.SubscribedSubjects {
			_, exist := m.SubscribeNodes[subject]
			if !exist {
				m.SubscribeNodes[subject] = []*node.Node{n}
				continue
			}

			m.SubscribeNodes[subject] = append(m.SubscribeNodes[subject], n)
		}
	}

	if len(n.BroadcastedSubjects) != 0 {
		for _, subject := range n.BroadcastedSubjects {
			_, exist := m.BroadCastNodes[subject]
			if !exist {
				m.BroadCastNodes[subject] = []*node.Node{n}
				continue
			}

			m.BroadCastNodes[subject] = append(m.BroadCastNodes[subject], n)
		}
	}

	return nil
}

// Removes a node from the Node Store.  Should be called when exiting a service
func (m *MockNodeStore) RemoveNode(n *node.Node) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(n.SubscribedSubjects) != 0 {
		for _, subject := range n.SubscribedSubjects {
			subscribers, exist := m.SubscribeNodes[subject]
			if !exist {
				continue
			}

			for i, s := range subscribers {
				if s.Id == n.Id {
					m.SubscribeNodes[subject][i] = m.SubscribeNodes[subject][len(m.SubscribeNodes[subject])-1]
					m.SubscribeNodes[subject] = m.SubscribeNodes[subject][:len(m.SubscribeNodes[subject])-1]
					break
				}
			}
		}
	}

	if len(n.BroadcastedSubjects) != 0 {
		for _, subject := range n.BroadcastedSubjects {
			broadcasters, exist := m.BroadCastNodes[subject]
			if !exist {
				continue
			}

			for i, b := range broadcasters {
				if b.Id == n.Id {
					m.BroadCastNodes[subject][i] = m.BroadCastNodes[subject][len(m.BroadCastNodes[subject])-1]
					m.BroadCastNodes[subject] = m.BroadCastNodes[subject][:len(m.BroadCastNodes[subject])-1]
					break
				}
			}
		}
	}

	return nil
}
