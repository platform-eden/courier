package courier

// interface that does the interactions with a Node Store
type NodeStorer interface {
	GetSubscribers(...string) ([]*Node, error)
	AddNode(*Node) error
	RemoveNode(*Node) error
}

type MockNodeStore struct {
	lock           ticketLock
	BroadCastNodes map[string][]*Node
	SubscribeNodes map[string][]*Node
}

// returns a mock of a Node Store good for testing
func NewMockNodeStore(nodes ...*Node) *MockNodeStore {
	bnodes := map[string][]*Node{}
	snodes := map[string][]*Node{}

	for _, node := range nodes {
		if len(node.SubscribedSubjects) != 0 {
			for _, subject := range node.SubscribedSubjects {
				_, exist := snodes[subject]
				if !exist {
					snodes[subject] = []*Node{node}
					continue
				}

				snodes[subject] = append(snodes[subject], node)
			}
		}

		if len(node.BroadcastedSubjects) != 0 {
			for _, subject := range node.BroadcastedSubjects {
				_, exist := bnodes[subject]
				if !exist {
					bnodes[subject] = []*Node{node}
					continue
				}

				bnodes[subject] = append(bnodes[subject], node)
			}
		}
	}

	m := MockNodeStore{
		lock:           *newTicketLock(),
		BroadCastNodes: bnodes,
		SubscribeNodes: snodes,
	}

	return &m
}

// Gets all of the subscribers that listen on the given subjects and returns them as a list
func (n *MockNodeStore) GetSubscribers(subjects ...string) ([]*Node, error) {
	n.lock.lock()
	defer n.lock.unlock()

	nodes := []*Node{}
	for _, subject := range subjects {
		subs, exist := n.SubscribeNodes[subject]
		if exist {
			nodes = append(nodes, subs...)
		}
	}
	return nodes, nil
}

// Adds a node to the Node Store.  Should be called when adding a courier service.
func (n *MockNodeStore) AddNode(node *Node) error {
	n.lock.lock()
	defer n.lock.unlock()

	if len(node.SubscribedSubjects) != 0 {
		for _, subject := range node.SubscribedSubjects {
			_, exist := n.SubscribeNodes[subject]
			if !exist {
				n.SubscribeNodes[subject] = []*Node{node}
				continue
			}

			n.SubscribeNodes[subject] = append(n.SubscribeNodes[subject], node)
		}
	}

	if len(node.BroadcastedSubjects) != 0 {
		for _, subject := range node.BroadcastedSubjects {
			_, exist := n.BroadCastNodes[subject]
			if !exist {
				n.BroadCastNodes[subject] = []*Node{node}
				continue
			}

			n.BroadCastNodes[subject] = append(n.BroadCastNodes[subject], node)
		}
	}

	return nil
}

// Removes a node from the Node Store.  Should be called when exiting a service
func (n *MockNodeStore) RemoveNode(node *Node) error {
	n.lock.lock()
	defer n.lock.unlock()

	if len(node.SubscribedSubjects) != 0 {
		for _, subject := range node.SubscribedSubjects {
			subscribers, exist := n.SubscribeNodes[subject]
			if !exist {
				continue
			}

			for i, s := range subscribers {
				if s.Id == node.Id {
					n.SubscribeNodes[subject][i] = n.SubscribeNodes[subject][len(n.SubscribeNodes[subject])-1]
					n.SubscribeNodes[subject] = n.SubscribeNodes[subject][:len(n.SubscribeNodes[subject])-1]
					break
				}
			}
		}
	}

	if len(node.BroadcastedSubjects) != 0 {
		for _, subject := range node.BroadcastedSubjects {
			broadcasters, exist := n.BroadCastNodes[subject]
			if !exist {
				continue
			}

			for i, b := range broadcasters {
				if b.Id == node.Id {
					n.BroadCastNodes[subject][i] = n.BroadCastNodes[subject][len(n.BroadCastNodes[subject])-1]
					n.BroadCastNodes[subject] = n.BroadCastNodes[subject][:len(n.BroadCastNodes[subject])-1]
					break
				}
			}
		}
	}

	return nil
}
