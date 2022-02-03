package client

import "github.com/platform-edn/courier/pkg/lock"

type clientNodeMap struct {
	nodes map[string]clientNode
	lock  lock.Locker
}

func newClientNodeMap() *clientNodeMap {
	nm := clientNodeMap{
		nodes: map[string]clientNode{},
		lock:  lock.NewTicketLock(),
	}

	return &nm
}

func (nm *clientNodeMap) Node(id string) (clientNode, bool) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	n, exist := nm.nodes[id]

	return n, exist
}

func (nm *clientNodeMap) AddClientNode(n clientNode) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	nm.nodes[n.Id] = n
}

func (nm *clientNodeMap) RemoveClientNode(id string) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	delete(nm.nodes, id)
}

func (nm *clientNodeMap) Length() int {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return len(nm.nodes)
}
